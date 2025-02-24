// Package soul holds the common types and functions used by the rest of the packages.
package soul

import (
	"errors"
	"math/rand/v2"
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

type Token uint32

func (t *Token) Gen() {
	n := rand.Uint32()
	*t = Token(n)
}

func (t Token) Uint32() uint32 {
	return uint32(t)
}

var ErrMismatchingCodes = errors.New("mismatching codes")

var ErrDifferentPacketSize = errors.New("the declared size of the package does not match the size of the actual read")

type ServerCode int

func (s ServerCode) String() string {
	switch s {
	case 1:
		return "Login"
	case 2:
		return "SetListenPort"
	case 3:
		return "GetPeerAddress"
	case 5:
		return "WatchUser"
	case 6:
		return "UnwatchUser"
	case 7:
		return "GetUserStatus"
	case 13:
		return "SayChatroom"
	case 14:
		return "JoinRoom"
	case 15:
		return "LeaveRoom"
	case 16:
		return "UserJoinedRoom"
	case 17:
		return "UserLeftRoom"
	case 18:
		return "ConnectToPeer"
	case 22:
		return "MessageUser"
	case 23:
		return "MessageAcked"
	case 26:
		return "FileSearch"
	case 28:
		return "SetStatus"
	case 32:
		return "ServerPing"
	case 35:
		return "SharedFoldersFiles"

	default:
		return "Unknown"
	}
}

type PeerInitCode int

type PeerCode int

type DistributedCode int
