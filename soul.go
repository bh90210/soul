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

const (
	// Peer connection type: Peer To Peer.
	Peer ConnectionType = "P"
	// File connection type: File Transfer.
	File ConnectionType = "F"
	// Distributed connection type: Distributed Network.
	Distributed ConnectionType = "D"
)

// ErrMismatchingCodes is returned when the code read from the stream does not match the expected of the consumer.
var ErrMismatchingCodes = errors.New("mismatching codes")

// ErrDifferentPacketSize is returned when the declared size of the package does not match the size of the actual read.
var ErrDifferentPacketSize = errors.New("the declared size of the package does not match the size of the actual read")

// ServerCode messages are used by clients to interface with the server over a connection (TCP).
type ServerCode int

// PeerInitCode This message is sent to initiate a direct connection to another peer.
type PeerInitCode int

// PeerCode messages are sent to peers over a P connection (TCP). Only a single
// active connection to a peer is allowed.
type PeerCode int

// DistributedCode messages are sent to peers over a D connection (TCP), and are used
// for the distributed search network. Only a single active connection to a peer is allowed.
type DistributedCode int

// Token is a unique identifier of type uint32 that is used throughout the protocol.
type Token uint32

// NewToken returns a new randomly generated uint32 long token.
// Under the hood, it uses math/rand/v2.Uint32().
func NewToken() Token {
	return Token(rand.Uint32())
}
