package streams

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"io"
	"net"
	"testing"
	"time"
)

func mustDecodeHex(t *testing.T, value string) []byte {
	t.Helper()

	decoded, err := hex.DecodeString(value)
	if err != nil {
		t.Fatalf("decoding %q: %v", value, err)
	}

	return decoded
}

// The cipher is only worth anything if it is the same one the client runs, so it
// is checked against the published example for AES-128 in 8-bit cipher feedback
// mode rather than against itself.
func TestCfb8MatchesTheStandardExample(t *testing.T) {
	key := mustDecodeHex(t, "2b7e151628aed2a6abf7158809cf4f3c")
	iv := mustDecodeHex(t, "000102030405060708090a0b0c0d0e0f")
	plaintext := mustDecodeHex(t, "6bc1bee22e409f96e93d7e117393172aae2d")
	ciphertext := mustDecodeHex(t, "3b79424c9c0dd436bace9e0ed4586a4f32b9")

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	encrypted := make([]byte, len(plaintext))
	newCfb8(block, iv, false).XORKeyStream(encrypted, plaintext)

	if !bytes.Equal(encrypted, ciphertext) {
		t.Errorf("encrypted % x, want % x", encrypted, ciphertext)
	}

	decrypted := make([]byte, len(ciphertext))
	newCfb8(block, iv, true).XORKeyStream(decrypted, ciphertext)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted % x, want % x", decrypted, plaintext)
	}
}

// A packet is written in pieces and read back in others, and the cipher carries
// its register across every one of them: a byte encrypted in one call has to
// decrypt the same as one encrypted in a call of its own.
func TestCfb8CarriesItsStateAcrossCalls(t *testing.T) {
	key := bytes.Repeat([]byte{0x2a}, 16)

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plaintext := []byte("a packet body long enough to cross a block boundary or three")

	encrypter := newCfb8(block, key, false)
	encrypted := make([]byte, 0, len(plaintext))

	for _, size := range []int{1, 7, 16, 17} {
		chunk := plaintext[len(encrypted):min(len(encrypted)+size, len(plaintext))]

		out := make([]byte, len(chunk))
		encrypter.XORKeyStream(out, chunk)
		encrypted = append(encrypted, out...)
	}

	out := make([]byte, len(plaintext)-len(encrypted))
	encrypter.XORKeyStream(out, plaintext[len(encrypted):])
	encrypted = append(encrypted, out...)

	decrypted := make([]byte, len(encrypted))
	newCfb8(block, key, true).XORKeyStream(decrypted, encrypted)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted %q, want %q", decrypted, plaintext)
	}

	// The secret is the initialization vector as well as the key, and the
	// register it becomes is rewritten for every byte. Taking it apart would
	// leave the other direction keyed to something else.
	if !bytes.Equal(key, bytes.Repeat([]byte{0x2a}, 16)) {
		t.Errorf("the shared secret is now % x, want it left alone", key)
	}
}

// A stream that was never on a connection has nothing to put the cipher
// underneath.
func TestEnableEncryptionRefusesAStreamWithoutAConnection(t *testing.T) {
	stream := NewMinecraftStreamFromBuffer(new(bytes.Buffer))

	if err := stream.EnableEncryption(bytes.Repeat([]byte{0x01}, 16)); err == nil {
		t.Error("error = nil, want a stream that is not on a connection refused")
	}
}

func TestEnableEncryptionRefusesASecretThatIsNotAKey(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	stream := NewMinecraftStreamFromNetConn(server)

	// Sixteen bytes is the only length the protocol sends, and it is the length
	// of the initialization vector as much as of the key.
	if err := stream.EnableEncryption([]byte("too short")); err == nil {
		t.Error("error = nil, want a secret of the wrong length refused")
	}

	if err := stream.EnableEncryption(bytes.Repeat([]byte{0x01}, 16)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Telling the connection twice would key one direction to a secret the other
	// end never had.
	if err := stream.EnableEncryption(bytes.Repeat([]byte{0x01}, 16)); err == nil {
		t.Error("error = nil, want encryption refused a second time")
	}
}

// The far end of a connection that has been told to encrypt, standing apart from
// the server's own cipher so that what is checked is that the two agree.
type encryptedPeer struct {
	conn      net.Conn
	encrypter *cfb8
	decrypter *cfb8
}

func newEncryptedPeer(t *testing.T, conn net.Conn, secret []byte) *encryptedPeer {
	t.Helper()

	block, err := aes.NewCipher(secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return &encryptedPeer{conn: conn, encrypter: newCfb8(block, secret, false), decrypter: newCfb8(block, secret, true)}
}

func (p *encryptedPeer) encrypt(plaintext []byte) []byte {
	encrypted := make([]byte, len(plaintext))
	p.encrypter.XORKeyStream(encrypted, plaintext)

	return encrypted
}

func (p *encryptedPeer) write(t *testing.T, plaintext []byte) {
	t.Helper()

	if _, err := p.conn.Write(p.encrypt(plaintext)); err != nil {
		t.Errorf("writing to the connection: %v", err)
	}
}

func (p *encryptedPeer) read(t *testing.T, size int) []byte {
	t.Helper()

	encrypted := make([]byte, size)
	if _, err := io.ReadFull(p.conn, encrypted); err != nil {
		t.Fatalf("reading from the connection: %v", err)
	}

	plaintext := make([]byte, size)
	p.decrypter.XORKeyStream(plaintext, encrypted)

	return plaintext
}

func TestEnableEncryptionEncryptsBothDirections(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	if err := client.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secret := bytes.Repeat([]byte{0x07}, 16)
	stream := NewMinecraftStreamFromNetConn(server)
	peer := newEncryptedPeer(t, client, secret)

	if err := stream.EnableEncryption(secret); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go func() {
		peer.write(t, []byte{0x05, 'h', 'e', 'l', 'l', 'o'})
	}()

	length, err := stream.ReadVarInt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, err := stream.ReadBytes(length)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(body) != "hello" {
		t.Errorf("read %q, want %q", body, "hello")
	}

	written := make(chan []byte, 1)
	go func() {
		written <- peer.read(t, 6)
	}()

	if err := stream.WriteString("world"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := <-written; !bytes.Equal(got, []byte{0x05, 'w', 'o', 'r', 'l', 'd'}) {
		t.Errorf("the connection carried % x, want the string encrypted", got)
	}
}

// A client that sends its next packet the instant it has answered leaves
// ciphertext in a buffer that was still reading plain. Whatever was pulled off
// the connection early has to be decrypted along with the rest, or the first
// frame after the cipher is a frame the connection never recovers from.
func TestEnableEncryptionDecryptsWhatWasBufferedTooEarly(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	if err := client.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secret := bytes.Repeat([]byte{0x07}, 16)
	stream := NewMinecraftStreamFromNetConn(server)
	peer := newEncryptedPeer(t, client, secret)

	// A plain byte and the first encrypted packet, arriving together. Reading
	// the plain one is what pulls the encrypted bytes in behind it, which is the
	// case this is about.
	go func() {
		if _, err := client.Write(append([]byte{0x2a}, peer.encrypt([]byte{0x05, 'h', 'e', 'l', 'l', 'o'})...)); err != nil {
			t.Errorf("writing to the connection: %v", err)
		}
	}()

	plain, err := stream.ReadByte()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plain != 0x2a {
		t.Fatalf("read %#02x, want the plain byte that came before the cipher", plain)
	}

	if buffered := stream.reader.Buffered(); buffered != 6 {
		t.Fatalf("the reader holds %d bytes it read early, want the 6 the packet behind the plain byte takes", buffered)
	}

	if err := stream.EnableEncryption(secret); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	length, err := stream.ReadVarInt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, err := stream.ReadBytes(length)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(body) != "hello" {
		t.Errorf("read %q, want the packet that arrived before the cipher was on", body)
	}
}

// Whatever is still waiting to be written was meant to travel plain, and
// flushing it after the swap would put it through the cipher instead.
func TestEnableEncryptionFlushesWhatWasWrittenPlain(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	if err := client.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secret := bytes.Repeat([]byte{0x07}, 16)
	stream := NewMinecraftStreamFromNetConn(server)

	read := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 4)
		if _, err := io.ReadFull(client, buf); err != nil {
			t.Errorf("reading from the connection: %v", err)
			close(read)

			return
		}

		read <- buf
	}()

	if err := stream.WriteBytes([]byte{0x01, 0x02, 0x03, 0x04}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := stream.EnableEncryption(secret); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := <-read; !bytes.Equal(got, []byte{0x01, 0x02, 0x03, 0x04}) {
		t.Errorf("the connection carried % x, want the bytes written before the cipher left plain", got)
	}
}
