// Package soul holds the common types and functions used by the rest of the packages.
package soul

import (
	"errors"
)

const (
	MajorVersion uint32 = 160
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

var ErrMismatchingCodes = errors.New("mismatching codes")

var ErrDifferentPacketSize = errors.New("the declared size of the package does not match the size of the actual read")

type ServerCode int

type PeerInitCode int

type PeerCode int

type DistributedCode int
