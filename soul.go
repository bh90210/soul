// Package soul holds the common types used by the rest of the sub-packages.
package soul

import (
	"errors"
	"math"
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

// Token is a unique identifier used throughout the protocol.
type Token uint32

// NewToken returns a new token and run t.Gen() on it before returning.
// It uses math/rand/v2 to generate the random uint32 token.
func NewToken() Token {
	t, _ := NewTokenFrom(rand.Uint32())
	return t
}

// ErrTokenNegative is returned when the token is negative.
var ErrTokenNegative = errors.New("token cannot be negative")

// ErrTokenTooLarge is returned when the token is greater than uint32.
var ErrTokenTooLarge = errors.New("token cannot be greater than uint32 (4294967295)")

type integer interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

// NewTokenFrom returns a new token from the given integer (int, int8, uin32, etc..)
// It returns an error if the integer is negative or greater than uint32.
func NewTokenFrom[I integer](i I) (Token, error) {
	if i < 0 {
		return 0, ErrTokenNegative
	}

	if any(i).(int) > math.MaxUint32 {
		return 0, ErrTokenTooLarge
	}

	return Token(i), nil
}
