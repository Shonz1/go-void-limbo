package streams

import (
	"bufio"
	"bytes"
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
}

func NewMinecraftStream(stream ReadWriter) *MinecraftStream {
	return &MinecraftStream{stream: stream}
}

func NewMinecraftStreamFromNetConn(conn net.Conn) *MinecraftStream {
	return NewMinecraftStream(bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)))
}

func NewMinecraftStreamFromBuffer(buf *bytes.Buffer) *MinecraftStream {
	return NewMinecraftStream(bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf)))
}

func (s *MinecraftStream) Flush() error {
	return s.stream.Flush()
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

func (s *MinecraftStream) ReadVarInt() (int32, error) {
	value := int32(0)

	for position := 0; position < 32; position += 7 {
		currentByte, err := s.ReadByte()
		if err != nil {
			return 0, err
		}

		value |= int32(currentByte&0x7F) << position

		if currentByte&0x80 == 0 {
			return value, nil
		}
	}

	return 0, errors.New("VarInt too big")
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
