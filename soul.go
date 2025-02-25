// Package soul holds the common types used by the rest of the sub-packages.
package soul

import (
	"errors"
	"math/rand/v2"
)

const (
	// MajorVersion is a known accepted client version accepted by the SoulSeek network.
	MajorVersion uint32 = 160
	// MinorVersion is a known accepted client version accepted by the SoulSeek network.
	MinorVersion uint32 = 0
)

// ConnectionType represents the type of connection.
type ConnectionType string

// ErrMismatchingCodes is returned when the code read from the stream does not match the expected of the consumer.
var ErrMismatchingCodes = errors.New("mismatching codes")

// ErrDifferentPacketSize is returned when the declared size of the package does not match the size of the actual read.
var ErrDifferentPacketSize = errors.New("the declared size of the package does not match the size of the actual read")

// CodeServer messages are used by clients to interface with the server over a connection (TCP).
type CodeServer int

// CodePeerInit This message is sent to initiate a direct connection to another peer.
type CodePeerInit int

// CodePeer messages are sent to peers over a P connection (TCP). Only a single
// active connection to a peer is allowed.
type CodePeer int

// CodeDistributed messages are sent to peers over a D connection (TCP), and are used
// for the distributed search network. Only a single active connection to a peer is allowed.
type CodeDistributed int

// Token is a unique identifier of type uint32 that is used throughout the protocol.
type Token uint32

// NewToken returns a new randomly generated uint32 long token.
// Under the hood, it uses math/rand/v2.Uint32().
func NewToken() Token {
	return Token(rand.Uint32())
}
