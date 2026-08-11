package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"time"
)

// The protocol version the bots handshake as. 776 is Minecraft 26.2, the latest
// version this limbo speaks, so a bot is a client the server has no reason to
// turn away.
const protocolVersion = 776

// The spawn the limbo teleports a joining player to, which is where a bot's
// position updates wander around. It has to match the server's own spawn so the
// coordinates stay inside the dimension.
const (
	spawnX = 0.5
	spawnY = 64.0
	spawnZ = 0.5
)

// The clientbound packet ids a bot has to recognise, grouped by the phase they
// arrive in. A limbo run for a load test has encryption off, so no encryption
// request and no plugin request ever reach a bot; the only login packets it sees
// are the two that end the login, plus a disconnect if something went wrong.
const (
	loginDisconnect  = 0x00
	loginSuccess     = 0x02
	loginSetCompress = 0x03

	configFinish    = 0x03
	configKeepAlive = 0x04

	playKeepAlive      = 0x2C
	playLogin          = 0x31
	playPlayerPosition = 0x48
)

// The serverbound packet ids a bot sends. These are the ids the limbo resolves
// each packet to on the latest protocol, and a bot speaks nothing else.
const (
	sbLoginStart        = 0x00
	sbLoginAcknowledged = 0x03

	sbAckFinishConfig = 0x03
	sbConfigKeepAlive = 0x04

	sbAcceptTeleport = 0x00
	sbMovePlayerPos  = 0x1E
	sbPlayKeepAlive  = 0x1C
)

type phase int

const (
	phaseLogin phase = iota
	phaseConfig
	phasePlay
)

// bot is one simulated player: the connection, the framing state it is on, and
// the phase it has reached. It drives the login the way a vanilla client does
// and then does the one thing a joined client has to keep doing, which is answer
// keep alives, so the server holds the connection open and counts it as online.
type bot struct {
	conn net.Conn
	r    *bufio.Reader

	// writeMu serialises the connection's write side. Once a bot is in play it is
	// written from two goroutines -- the read loop answering keep alives and
	// teleports, and the mover sending position updates -- and a frame from one
	// must not interleave with a frame from the other.
	writeMu sync.Mutex

	// compressed is set once the server announces a compression threshold, from
	// which point every frame carries a data-length prefix.
	compressed bool

	// movesPerSecond is how many position updates the bot streams once it has
	// joined, which is the load this test is really about.
	movesPerSecond int

	phase phase
}

// dialBot opens a connection and plays the handshake and login start, the two
// packets a client always sends in the clear before anything is negotiated.
func dialBot(addr, host string, port uint16, username, uuid string, movesPerSecond int) (*bot, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	b := &bot{conn: conn, r: bufio.NewReaderSize(conn, 4096), phase: phaseLogin, movesPerSecond: movesPerSecond}

	// Handshake: protocol version, the address the client reached the server at,
	// the port, and intent 2 for login.
	handshake := new(bytes.Buffer)
	writeVarInt(handshake, protocolVersion)
	writeString(handshake, host)
	binary.Write(handshake, binary.BigEndian, port)
	writeVarInt(handshake, 2)

	if err := b.write(0x00, handshake.Bytes()); err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake: %w", err)
	}

	// Login start: the name the bot claims and a uuid it made up. With encryption
	// off the server takes the name at its word, so both are only ever echoed
	// back.
	loginStart := new(bytes.Buffer)
	writeString(loginStart, username)
	loginStart.Write(uuidBytes(uuid))

	if err := b.write(sbLoginStart, loginStart.Bytes()); err != nil {
		conn.Close()
		return nil, fmt.Errorf("login start: %w", err)
	}

	return b, nil
}

// run drives the connection until it ends: it reads packets, answers the ones
// that keep the login moving and then the keep alives that keep it alive, and
// returns when the connection is closed or stop is signalled. It is the whole of
// a bot's life, and does nothing a real client would not.
func (b *bot) run(stop <-chan struct{}, onJoin func()) error {
	// A goroutine that closes the connection when stop fires, which unblocks the
	// read below so the bot can go away with the rest of them.
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-stop:
			b.conn.Close()
		case <-done:
		}
	}()

	announcedJoin := false

	for {
		id, body, err := b.read()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}

		switch b.phase {
		case phaseLogin:
			switch id {
			case loginSetCompress:
				threshold, _, _ := readVarInt(bytes.NewReader(body))
				b.compressed = threshold >= 0
			case loginSuccess:
				// Acknowledge the success, which is what moves both ends into
				// configuration.
				if err := b.write(sbLoginAcknowledged, nil); err != nil {
					return err
				}
				b.phase = phaseConfig
			case loginDisconnect:
				return fmt.Errorf("disconnected during login")
			}

		case phaseConfig:
			switch id {
			case configKeepAlive:
				if err := b.write(sbConfigKeepAlive, body); err != nil {
					return err
				}
			case configFinish:
				if err := b.write(sbAckFinishConfig, nil); err != nil {
					return err
				}
				b.phase = phasePlay

				// The join is done, so the mover starts and streams position
				// updates until the connection ends, which is what done signals.
				if b.movesPerSecond > 0 {
					go b.sendPositions(done)
				}
			}

		case phasePlay:
			switch id {
			case playKeepAlive:
				// The answer is the same eight bytes the server sent, which is the
				// whole of what proves the bot is still there.
				if err := b.write(sbPlayKeepAlive, body); err != nil {
					return err
				}
			case playPlayerPosition:
				teleportId, _, _ := readVarInt(bytes.NewReader(body))
				ack := new(bytes.Buffer)
				writeVarInt(ack, teleportId)
				if err := b.write(sbAcceptTeleport, ack.Bytes()); err != nil {
					return err
				}
			case playLogin:
				if !announcedJoin && onJoin != nil {
					announcedJoin = true
					onJoin()
				}
			}
		}
	}
}

// sendPositions streams movesPerSecond position updates a second until done is
// closed, which is a real client's busiest serverbound traffic and the load
// this test exists to put on the server. The position drifts by a hair each
// tick so the packets are not identical, though a limbo reads none of the
// numbers -- the cost it measures is the framing, the decode and the goroutine
// wakeup, which are the same whatever the coordinates say.
func (b *bot) sendPositions(done <-chan struct{}) {
	ticker := time.NewTicker(time.Second / time.Duration(b.movesPerSecond))
	defer ticker.Stop()

	tick := 0

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			tick++

			// A small wander around the spawn the server placed the bot at, kept
			// well inside the dimension so nothing about the coordinates is what
			// ends the connection.
			offset := float64(tick%20) * 0.05

			body := new(bytes.Buffer)
			writeDouble(body, spawnX+offset)
			writeDouble(body, spawnY)
			writeDouble(body, spawnZ+offset)
			body.WriteByte(0x00) // neither on the ground nor colliding

			if err := b.write(sbMovePlayerPos, body.Bytes()); err != nil {
				return
			}
		}
	}
}

// write frames a packet body under its id and sends it, adding the compressed
// framing once the connection is on it. Every packet a bot sends is far smaller
// than any threshold, so it always travels in full with a zero data length,
// which is the one branch of compression a bot ever needs to produce. It holds
// writeMu because a joined bot writes from both its read loop and its mover.
func (b *bot) write(id int32, payload []byte) error {
	b.writeMu.Lock()
	defer b.writeMu.Unlock()

	body := new(bytes.Buffer)
	writeVarInt(body, id)
	body.Write(payload)

	frame := new(bytes.Buffer)

	if b.compressed {
		inner := new(bytes.Buffer)
		writeVarInt(inner, 0) // data length zero: sent uncompressed
		inner.Write(body.Bytes())
		writeVarInt(frame, int32(inner.Len()))
		frame.Write(inner.Bytes())
	} else {
		writeVarInt(frame, int32(body.Len()))
		frame.Write(body.Bytes())
	}

	if err := b.conn.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return err
	}

	_, err := b.conn.Write(frame.Bytes())
	return err
}

// read pulls one frame off the connection and returns the packet id and body
// inside it, inflating the body when the frame carried a non-zero data length.
func (b *bot) read() (int32, []byte, error) {
	if err := b.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		return 0, nil, err
	}

	length, _, err := readVarInt(b.r)
	if err != nil {
		return 0, nil, err
	}

	if length < 1 {
		return 0, nil, fmt.Errorf("invalid frame length %d", length)
	}

	raw := make([]byte, length)
	if _, err := io.ReadFull(b.r, raw); err != nil {
		return 0, nil, err
	}

	data := raw

	if b.compressed {
		size, n, err := readVarIntBytes(raw)
		if err != nil {
			return 0, nil, err
		}

		payload := raw[n:]

		if size != 0 {
			inflated, err := inflate(payload, int(size))
			if err != nil {
				return 0, nil, err
			}
			data = inflated
		} else {
			data = payload
		}
	}

	id, n, err := readVarIntBytes(data)
	if err != nil {
		return 0, nil, err
	}

	return id, data[n:], nil
}

// inflate undoes the server's zlib compression on a body big enough to have been
// worth compressing.
func inflate(data []byte, size int) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer zr.Close()

	out := make([]byte, 0, size)
	buf := bytes.NewBuffer(out)

	if _, err := io.Copy(buf, zr); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// --- wire helpers, kept apart from the server's own so a bot exercises the
// protocol rather than the code under test. ---

func writeVarInt(buf *bytes.Buffer, value int32) {
	u := uint32(value)
	for {
		b := byte(u & 0x7F)
		u >>= 7
		if u != 0 {
			b |= 0x80
		}
		buf.WriteByte(b)
		if u == 0 {
			return
		}
	}
}

func writeString(buf *bytes.Buffer, value string) {
	writeVarInt(buf, int32(len(value)))
	buf.WriteString(value)
}

func writeDouble(buf *bytes.Buffer, value float64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], math.Float64bits(value))
	buf.Write(b[:])
}

func readVarInt(r io.ByteReader) (int32, int, error) {
	var value uint32
	var read int
	for shift := 0; shift < 35; shift += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, read, err
		}
		read++
		value |= uint32(b&0x7F) << shift
		if b&0x80 == 0 {
			return int32(value), read, nil
		}
	}
	return 0, read, fmt.Errorf("var int too long")
}

func readVarIntBytes(b []byte) (int32, int, error) {
	return readVarInt(bytes.NewReader(b))
}

// uuidBytes turns a hyphenated uuid into the sixteen bytes the protocol carries.
// The bot invents its own, so a malformed one would be a bug here rather than
// anything the server said, and the zero uuid is a fine fallback.
func uuidBytes(uuid string) []byte {
	out := make([]byte, 16)

	nibble := 0
	for i := 0; i < len(uuid) && nibble < 32; i++ {
		c := uuid[i]
		if c == '-' {
			continue
		}

		var v byte
		switch {
		case c >= '0' && c <= '9':
			v = c - '0'
		case c >= 'a' && c <= 'f':
			v = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			v = c - 'A' + 10
		default:
			continue
		}

		if nibble%2 == 0 {
			out[nibble/2] = v << 4
		} else {
			out[nibble/2] |= v
		}
		nibble++
	}

	return out
}
