// Package testutil holds the test doubles the client and server tests share:
// the peer side of a connection, the client side of the cipher and the
// compressed framing, and the session server and proxy a real login would
// involve. Everything here stands apart from the server's own implementations
// on purpose, so that what the tests check is that the two ends agree.
package testutil

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/Shonz1/go-void-limbo/auth"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// FakeSessionServer stands in for Mojang's, recording what it was asked.
type FakeSessionServer struct {
	Usernames []string
	Hashes    []string

	Profile types.GameProfile
	Err     error
}

func (s *FakeSessionServer) HasJoined(username, serverHash string) (types.GameProfile, error) {
	s.Usernames = append(s.Usernames, username)
	s.Hashes = append(s.Hashes, serverHash)

	return s.Profile, s.Err
}

// KeyPair is generated once for the whole test run. Generating one is the
// slowest thing in these tests and none of them care which key they get.
var KeyPair = sync.OnceValue(func() *auth.KeyPair {
	keyPair, err := auth.NewKeyPair()
	if err != nil {
		panic(err)
	}

	return keyPair
})

// EncryptForServer is what a client does to the two fields of its encryption
// response: each one under the server's public key, with the padding the
// client's cipher uses.
func EncryptForServer(t *testing.T, plaintext []byte) []byte {
	t.Helper()

	parsed, err := x509.ParsePKIXPublicKey(KeyPair().PublicKey())
	if err != nil {
		t.Fatalf("reading the server key: %v", err)
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, parsed.(*rsa.PublicKey), plaintext)
	if err != nil {
		t.Fatalf("encrypting for the server: %v", err)
	}

	return ciphertext
}

// Cipher is the client's side of the connection cipher: AES in 8-bit cipher
// feedback mode, keyed by the shared secret and started from it. It stands apart
// from the server's own so that what is checked is that the two agree.
type Cipher struct {
	block     cipher.Block
	register  []byte
	keyStream []byte
	decrypt   bool
}

func NewCipher(t *testing.T, secret []byte, decrypt bool) *Cipher {
	t.Helper()

	block, err := aes.NewCipher(secret)
	if err != nil {
		t.Fatalf("keying the cipher: %v", err)
	}

	register := make([]byte, len(secret))
	copy(register, secret)

	return &Cipher{block: block, register: register, keyStream: make([]byte, block.BlockSize()), decrypt: decrypt}
}

func (c *Cipher) Apply(data []byte) []byte {
	out := make([]byte, len(data))

	for i, in := range data {
		c.block.Encrypt(c.keyStream, c.register)

		out[i] = in ^ c.keyStream[0]

		feedback := out[i]
		if c.decrypt {
			feedback = in
		}

		copy(c.register, c.register[1:])
		c.register[len(c.register)-1] = feedback
	}

	return out
}

func DecryptFromServer(t *testing.T, secret, ciphertext []byte) []byte {
	t.Helper()

	return NewCipher(t, secret, true).Apply(ciphertext)
}

// Deflate is what a client that was told a threshold does to a body big enough
// for it. It stands apart from the server's own compression so that a frame
// built here is a frame the other end would have built.
func Deflate(t *testing.T, body []byte) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	writer := zlib.NewWriter(buf)

	if _, err := writer.Write(body); err != nil {
		t.Fatalf("compressing the body: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("compressing the body: %v", err)
	}

	return buf.Bytes()
}

// Frame puts the length in front of a packet body, which is the whole of the
// framing on a connection that has not been told a threshold.
func Frame(t *testing.T, body []byte) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := stream.WriteVarInt(int32(len(body))); err != nil {
		t.Fatalf("framing the body: %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("framing the body: %v", err)
	}

	return append(buf.Bytes(), body...)
}

// CompressedFrame is Frame for a connection that was told a threshold: the body
// gains a var int saying what it inflates to, or zero when it was small enough
// to travel in full. dataLength is written as given so that a client behaving
// badly can be framed too.
func CompressedFrame(t *testing.T, body []byte, dataLength int32) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := stream.WriteVarInt(dataLength); err != nil {
		t.Fatalf("framing the body: %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("framing the body: %v", err)
	}

	if dataLength != 0 {
		body = Deflate(t, body)
	}

	return Frame(t, append(buf.Bytes(), body...))
}

// Inflate undoes Deflate, for reading back what the server framed.
func Inflate(t *testing.T, data []byte) []byte {
	t.Helper()

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("inflating the body: %v", err)
	}

	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("inflating the body: %v", err)
	}

	return body
}

// FramedBody splits a frame written to a compressed connection into the size
// the body inflates to, which is zero for one left in full, and the body
// itself. It fails the test on a frame whose length does not match what
// follows it, since a client reads the next frame from where this one ends.
func FramedBody(t *testing.T, written []byte) (int32, []byte) {
	t.Helper()

	length, read, err := streams.ReadVarIntFrom(written)
	if err != nil {
		t.Fatalf("reading the packet length: %v", err)
	}

	body := written[read:]
	if int(length) != len(body) {
		t.Fatalf("length says %d bytes, frame carries %d", length, len(body))
	}

	dataLength, read, err := streams.ReadVarIntFrom(body)
	if err != nil {
		t.Fatalf("reading the data length: %v", err)
	}

	return dataLength, body[read:]
}

// IdAndBody splits one packet written to an uncompressed connection into the id
// in front of it and the body behind it. The two come from different places: the
// id is resolved at the version the client speaks, and the body is what the
// transformers left behind on their way down to it.
func IdAndBody(t *testing.T, written []byte) (types.PacketId, []byte) {
	t.Helper()

	length, read, err := streams.ReadVarIntFrom(written)
	if err != nil {
		t.Fatalf("reading the packet length: %v", err)
	}

	body := written[read:]
	if int(length) != len(body) {
		t.Fatalf("length says %d bytes, frame carries %d", length, len(body))
	}

	packetId, read, err := streams.ReadVarIntFrom(body)
	if err != nil {
		t.Fatalf("reading the packet id: %v", err)
	}

	return packetId, body[read:]
}

// LoginPeer is the far side of a login: it sends what a client sends and reads
// what a client reads, taking on the cipher and then the compressed framing at
// the points a real client takes them on.
type LoginPeer struct {
	T    *testing.T
	Conn net.Conn

	// Compressed says the peer frames everything for a threshold from here on,
	// which is what a client does the moment it reads set compression.
	Compressed bool

	encrypter *Cipher
	decrypter *Cipher
}

// Encrypt turns the connection cipher on, as a client does the instant it has
// sent its encryption response.
func (p *LoginPeer) Encrypt(secret []byte) {
	p.encrypter = NewCipher(p.T, secret, false)
	p.decrypter = NewCipher(p.T, secret, true)
}

func (p *LoginPeer) WritePacket(body []byte) {
	p.T.Helper()

	// Everything this side sends is small enough to travel in full, so a zero
	// data length is the whole of what compression adds to it.
	if p.Compressed {
		body = append([]byte{0x00}, body...)
	}

	written := Frame(p.T, body)
	if p.encrypter != nil {
		written = p.encrypter.Apply(written)
	}

	if _, err := p.Conn.Write(written); err != nil {
		p.T.Fatalf("writing to the connection: %v", err)
	}
}

func (p *LoginPeer) readByte() byte {
	p.T.Helper()

	buf := make([]byte, 1)
	if _, err := io.ReadFull(p.Conn, buf); err != nil {
		p.T.Fatalf("reading from the connection: %v", err)
	}

	if p.decrypter != nil {
		buf = p.decrypter.Apply(buf)
	}

	return buf[0]
}

func (p *LoginPeer) readVarInt() int32 {
	p.T.Helper()

	value := int32(0)

	for position := 0; position < 32; position += 7 {
		current := p.readByte()
		value |= int32(current&0x7F) << position

		if current&0x80 == 0 {
			return value
		}
	}

	p.T.Fatal("reading from the connection: var int too big")

	return 0
}

// ReadPacket reads one frame and returns the body inside it, through whichever
// of the cipher and the compressed framing the connection has reached.
func (p *LoginPeer) ReadPacket() *streams.MinecraftStream {
	p.T.Helper()

	length := p.readVarInt()

	body := make([]byte, length)
	for i := range body {
		body[i] = p.readByte()
	}

	if p.Compressed {
		size, read, err := streams.ReadVarIntFrom(body)
		if err != nil {
			p.T.Fatalf("reading the data length: %v", err)
		}

		body = body[read:]
		if size != 0 {
			body = Inflate(p.T, body)
		}
	}

	return streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))
}

// ForwardingSecret is what an operator shares between a modern proxy and this
// server, and is the whole of what makes a forwarded login worth anything.
var ForwardingSecret = []byte("a shared secret")

// SignedForwardingPayload builds the answer a modern proxy gives: the fields of
// the login, with the digest it takes over them under the shared secret in
// front.
func SignedForwardingPayload(t *testing.T, secret []byte, address, uuid, username string, properties []types.ProfileProperty) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	writes := []func() error{
		func() error { return stream.WriteVarInt(auth.ModernForwardingVersion) },
		func() error { return stream.WriteString(address) },
		func() error { return stream.WriteUuid(uuid) },
		func() error { return stream.WriteString(username) },
		func() error { return stream.WriteVarInt(int32(len(properties))) },
	}

	for _, property := range properties {
		writes = append(writes,
			func() error { return stream.WriteString(property.Name) },
			func() error { return stream.WriteString(property.Value) },
			func() error { return stream.WriteBoolean(property.Signature != nil) },
		)

		if property.Signature != nil {
			writes = append(writes, func() error { return stream.WriteString(*property.Signature) })
		}
	}

	writes = append(writes, stream.Flush)

	for _, write := range writes {
		if err := write(); err != nil {
			t.Fatalf("building the forwarding payload: %v", err)
		}
	}

	digest := hmac.New(sha256.New, secret)
	digest.Write(buf.Bytes())

	return append(digest.Sum(nil), buf.Bytes()...)
}
