package server

import (
	"sync/atomic"

	"github.com/Shonz1/go-void-limbo/types"
)

// playerCount is how many clients have reached the play phase, which is the
// number a server list ping is answered with.
//
// One is shared by every connection a server accepts: a ping asks about the
// server rather than about the connection it arrived on. It is read while a ping
// is being answered and written as connections join and leave, each on a
// goroutine of its own, so it is atomic rather than under any one connection's
// lock.
type playerCount struct {
	count atomic.Int64
}

// join counts a client that has just reached the play phase, leave stops
// counting one whose connection has ended, and online is what a ping reports.
func (p *playerCount) join() {
	p.count.Add(1)
}

func (p *playerCount) leave() {
	p.count.Add(-1)
}

func (p *playerCount) online() int32 {
	return int32(p.count.Load())
}

// status is what a server list ping is answered from: what the operator set the
// server to say about itself, and the count every connection joins and leaves.
//
// All of it belongs to the server rather than to any one connection, so there is
// one of these per server and a connection holds a pointer to it, through
// client.StatusProvider. That is what keeps a field added here from being
// copied onto every connection the server accepts.
type status struct {
	description string
	players     playerCount
}

// Status assembles what a ping arriving on a connection speaking version is
// answered with. The version is the only part of the answer the connection has
// any say in, which is why it is the only thing passed in.
func (s *status) Status(version types.ProtocolVersion) types.ServerStatus {
	online := s.players.online()

	return types.ServerStatus{
		Version: statusVersion(version),
		Players: types.ServerPlayers{
			Online: online,

			// A limbo turns nobody away, and the protocol has no way of saying
			// so: the field is a number, and a client draws a server as full when
			// the two are equal. One more than however many are on is the closest
			// this can come to the truth, and it is a truth about this server
			// rather than a number an operator was asked to invent.
			Max: online + 1,
		},
		Description: types.TextComponent{Text: s.description},
	}
}

// PlayerJoined and PlayerLeft are the two sides of the count: a client arriving
// in the play phase, and one whose connection has ended leaving it.
func (s *status) PlayerJoined() {
	s.players.join()
}

func (s *status) PlayerLeft() {
	s.players.leave()
}

// statusVersion is the version a ping is answered with: the client's own
// whenever this server speaks it, so that a client on any supported version sees
// a server it can join rather than one number it has to call incompatible.
//
// A client on a version this server does not speak left the connection on
// protocol zero, and is told the latest instead. That is the answer it came for:
// it has no use for a version it cannot join at, only for something to draw
// beside the fact that it cannot.
//
// Being on the chain is not the same as having a name to be drawn under, so the
// name is taken only if there is one. A version with none would otherwise panic
// here, which on the one handler a connection reaches without logging in would
// take the process with it.
func statusVersion(version types.ProtocolVersion) types.ServerVersion {
	if !types.IsSupportedProtocolVersion(version) || len(version.Names) == 0 {
		version = types.LatestProtocolVersion
	}

	name := ""
	if len(version.Names) > 0 {
		name = version.Names[0]
	}

	return types.ServerVersion{Name: name, Protocol: version.ID}
}
