package streams

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"strings"
)

type ReadWriter interface {
	ReadByte() (byte, error)
	Read(b []byte) (n int, err error)
	WriteByte(c byte) error
	Write(b []byte) (n int, err error)
	Flush() error
}

type MinecraftStream struct {
	stream ReadWriter

	// conn is the connection the buffering sits on, for the streams that have
	// one. Encryption goes underneath the buffering rather than above it, so
	// turning it on means building the buffering again on top of the cipher,
	// which is only possible while the connection itself is still at hand.
	conn io.ReadWriter

	// reader is the read half of stream, kept apart from it because what it
	// pulled off the connection ahead of the cipher being turned on is
	// ciphertext it read as plain, and only the reader itself can say how much
	// of that there is.
	reader *bufio.Reader

	encrypted bool
}

func NewMinecraftStream(stream ReadWriter) *MinecraftStream {
	return &MinecraftStream{stream: stream}
}

func NewMinecraftStreamFromNetConn(conn net.Conn) *MinecraftStream {
	reader := bufio.NewReader(conn)

	return &MinecraftStream{
		stream: bufio.NewReadWriter(reader, bufio.NewWriter(conn)),
		conn:   conn,
		reader: reader,
	}
}

func NewMinecraftStreamFromBuffer(buf *bytes.Buffer) *MinecraftStream {
	return NewMinecraftStream(bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf)))
}

func (s *MinecraftStream) Flush() error {
	return s.stream.Flush()
}

// EnableEncryption puts every byte the connection carries from here on, in both
// directions, under AES keyed by the shared secret the client sent. The secret
// is its own initialization vector, which is what the protocol asks for and what
// makes the two directions symmetric.
//
// The cipher goes underneath the buffering, so that everything above it goes on
// reading and writing plaintext. What the buffered reader had already pulled off
// the connection is the one thing that does not fit that: those bytes were read
// as plain and are ciphertext, so they are decrypted here and put back in front
// of the connection. A client has nothing to send between its encryption
// response and the reply to it, so there is rarely anything there, but a frame
// lost to a read that came a moment early is a connection that never recovers.
func (s *MinecraftStream) EnableEncryption(secret []byte) error {
	if s.conn == nil {
		return errors.New("the stream is not on a connection")
	}

	if s.encrypted {
		return errors.New("encryption is already enabled")
	}

	block, err := aes.NewCipher(secret)
	if err != nil {
		return fmt.Errorf("invalid shared secret: %w", err)
	}

	if len(secret) != block.BlockSize() {
		return fmt.Errorf("shared secret is %d bytes, which is not the %d byte initialization vector the cipher needs", len(secret), block.BlockSize())
	}

	// Anything still waiting to be written was meant to travel plain, and
	// flushing it after the swap would send it through the cipher instead.
	if err := s.stream.Flush(); err != nil {
		return err
	}

	pending, err := s.reader.Peek(s.reader.Buffered())
	if err != nil {
		return err
	}

	decrypter := newCfb8(block, secret, true)

	buffered := make([]byte, len(pending))
	decrypter.XORKeyStream(buffered, pending)

	reader := bufio.NewReader(io.MultiReader(bytes.NewReader(buffered), cipher.StreamReader{S: decrypter, R: s.conn}))
	writer := bufio.NewWriter(cipher.StreamWriter{S: newCfb8(block, secret, false), W: s.conn})

	s.reader = reader
	s.stream = bufio.NewReadWriter(reader, writer)
	s.encrypted = true

	return nil
}

func (s *MinecraftStream) ReadByte() (byte, error) {
	return s.stream.ReadByte()
}

func (s *MinecraftStream) WriteByte(b byte) error {
	return s.stream.WriteByte(b)
}

func (s *MinecraftStream) ReadBytes(size int32) ([]byte, error) {
	buf := make([]byte, size)
	_, err := io.ReadFull(s.stream, buf)
	return buf, err
}

func (s *MinecraftStream) WriteBytes(b []byte) error {
	_, err := s.stream.Write(b)
	return err
}

// ReadByteArray reads a var int length and then that many bytes. The length
// arrives from the other end, and the buffer for it is allocated before a single
// byte of it has been read, so max is what the caller knows the field can hold:
// anything longer is refused rather than reserved for.
func (s *MinecraftStream) ReadByteArray(max int32) ([]byte, error) {
	length, err := s.ReadVarInt()
	if err != nil {
		return nil, err
	}

	if length < 0 || length > max {
		return nil, fmt.Errorf("invalid byte array length: %d", length)
	}

	return s.ReadBytes(length)
}

// WriteByteArray writes a var int length followed by the bytes themselves.
func (s *MinecraftStream) WriteByteArray(b []byte) error {
	if err := s.WriteVarInt(int32(len(b))); err != nil {
		return err
	}

	return s.WriteBytes(b)
}

func (s *MinecraftStream) ReadVarInt() (int32, error) {
	value, _, err := readVarInt(s.stream)
	return value, err
}

// ReadVarIntFrom reads a var int from the front of b and reports how many bytes
// it took up. A compressed packet has to be split at the end of the var int in
// front of its body, and a buffered reader cannot say where that is.
func ReadVarIntFrom(b []byte) (int32, int, error) {
	return readVarInt(bytes.NewReader(b))
}

// readVarInt reads a var int a byte at a time, returning it along with the
// number of bytes it occupied.
func readVarInt(reader io.ByteReader) (int32, int, error) {
	value := int32(0)
	size := 0

	for position := 0; position < 32; position += 7 {
		currentByte, err := reader.ReadByte()
		if err != nil {
			return 0, size, err
		}

		size++
		value |= int32(currentByte&0x7F) << position

		if currentByte&0x80 == 0 {
			return value, size, nil
		}
	}

	return 0, size, errors.New("VarInt too big")
}

func (s *MinecraftStream) WriteVarInt(value int32) error {
	uvalue := uint32(value)
	for (uvalue & ^uint32(0x7F)) != 0 {
		err := s.WriteByte(byte((uvalue & 0x7F) | 0x80))
		if err != nil {
			return err
		}
		uvalue >>= 7
	}

	return s.WriteByte(byte(uvalue))
}

func (s *MinecraftStream) ReadString() (string, error) {
	length, err := s.ReadVarInt()
	if err != nil {
		return "", err
	}

	value, err := s.ReadBytes(length)
	if err != nil {
		return "", err
	}

	return string(value), nil
}

func (s *MinecraftStream) WriteString(value string) error {
	length := int32(len(value))
	err := s.WriteVarInt(length)
	if err != nil {
		return err
	}

	return s.WriteBytes([]byte(value))
}

func (s *MinecraftStream) ReadShort() (int16, error) {
	bytes, err := s.ReadBytes(2)
	if err != nil {
		return 0, err
	}

	return int16(binary.BigEndian.Uint16(bytes)), nil
}

func (s *MinecraftStream) WriteShort(value int16) error {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, uint16(value))
	return s.WriteBytes(bytes)
}

func (s *MinecraftStream) ReadInt() (int32, error) {
	bytes, err := s.ReadBytes(4)
	if err != nil {
		return 0, err
	}

	return int32(binary.BigEndian.Uint32(bytes)), nil
}

func (s *MinecraftStream) WriteInt(value int32) error {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(value))
	return s.WriteBytes(bytes)
}

func (s *MinecraftStream) ReadLong() (int64, error) {
	bytes, err := s.ReadBytes(8)
	if err != nil {
		return 0, err
	}

	return int64(binary.BigEndian.Uint64(bytes)), nil
}

func (s *MinecraftStream) WriteLong(value int64) error {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(value))
	return s.WriteBytes(bytes)
}

func (s *MinecraftStream) ReadFloat() (float32, error) {
	bits, err := s.ReadInt()
	if err != nil {
		return 0, err
	}

	return math.Float32frombits(uint32(bits)), nil
}

func (s *MinecraftStream) WriteFloat(value float32) error {
	return s.WriteInt(int32(math.Float32bits(value)))
}

func (s *MinecraftStream) ReadDouble() (float64, error) {
	bits, err := s.ReadLong()
	if err != nil {
		return 0, err
	}

	return math.Float64frombits(uint64(bits)), nil
}

func (s *MinecraftStream) WriteDouble(value float64) error {
	return s.WriteLong(int64(math.Float64bits(value)))
}

func (s *MinecraftStream) ReadBoolean() (bool, error) {
	value, err := s.ReadByte()
	if err != nil {
		return false, err
	}

	return value != 0, nil
}

func (s *MinecraftStream) WriteBoolean(value bool) error {
	if value {
		return s.WriteByte(1)
	}

	return s.WriteByte(0)
}

func (s *MinecraftStream) ReadUuid() (string, error) {
	long1, err := s.ReadLong()
	if err != nil {
		return "", err
	}

	long2, err := s.ReadLong()
	if err != nil {
		return "", err
	}

	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[0:8], uint64(long1))
	binary.BigEndian.PutUint64(buf[8:16], uint64(long2))

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16]), nil
}

// WriteUuid writes a hyphenated UUID string as the two big-endian longs the
// protocol expects.
func (s *MinecraftStream) WriteUuid(value string) error {
	buf, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	if err != nil {
		return fmt.Errorf("invalid uuid %q: %w", value, err)
	}

	if len(buf) != 16 {
		return fmt.Errorf("invalid uuid %q: expected 16 bytes, got %d", value, len(buf))
	}

	return s.WriteBytes(buf)
}
