// Package soul holds the common types used by the rest of the sub-packages.
package soul

import (
	"errors"
	"math/rand/v2"
)

// ConnectionType represents the type of connection. Possible values are "P", "F" and "D".
type ConnectionType string

// ErrMismatchingCodes is returned when the code read from the stream does not match the expected of the consumer.
var ErrMismatchingCodes = errors.New("mismatching codes")

// ErrDifferentPacketSize is returned when the declared size of the package does not match the size of the actual read.
var ErrDifferentPacketSize = errors.New("the declared size of the package does not match the size of the actual read")

// Token is a unique identifier of type uint32 that is used throughout the protocol.
type Token uint32

// NewToken returns a new randomly generated uint32 long token.
// Under the hood, it uses math/rand/v2.Uint32().
func NewToken() Token {
	return Token(rand.Uint32())
}
