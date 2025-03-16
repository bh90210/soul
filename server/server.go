// Package server messages are used by clients to interface with the server
// over a connection (TCP).
package server

//go:generate stringer -type Code -trimprefix Code
//go:generate stringer -type UserStatus -trimprefix Status

import (
	"bytes"
	"net"

	"github.com/bh90210/soul/internal"
)

// Code represents the type of server message.
type Code int

// UserStatus represents the status of a user.
type UserStatus int

const (
	// StatusOffline user status.
	StatusOffline UserStatus = iota
	// StatusAway user status.
	StatusAway
	// StatusOnline user status.
	StatusOnline
)

// Read reads a message from a server connection. It reads the size of the message
// and the code of the message. It then reads the message from the connection and
// returns the message, the size of the message, the code of the message and an error.
func Read(connection net.Conn) (*bytes.Buffer, int, Code, error) {
	r, s, c, err := internal.MessageRead(internal.CodeServer(0), connection, false)
	return r, s, Code(c), err
}

type message[M any] interface {
	*AcceptChildren |
		*BranchLevel |
		*BranchRoot |
		*CantConnectToPeer |
		*ChangePassword |
		*CheckPrivileges |
		*ConnectToPeer |
		*FileSearch |
		*GetPeerAddress |
		*GetUserStats |
		*GetUserStatus |
		*HaveNoParent |
		*JoinRoom |
		*LeaveRoom |
		*Login |
		*MessageAcked |
		*MessageUser |
		*MessageUsers |
		*Ping |
		*PrivateRoomAddOperator |
		*PrivateRoomAddUser |
		*PrivateRoomCancelMembership |
		*PrivateRoomDisown |
		*PrivateRoomRemoveOperator |
		*PrivateRoomRemoveUser |
		*PrivateRoomToggle |
		*RoomList |
		*RoomSearch |
		*RoomTickerSet |
		*SayChatroom |
		*SendUploadSpeed |
		*SetListenPort |
		*SetStatus |
		*SharedFoldersFiles |
		*UnwatchUser |
		*UserSearch |
		*WatchUser |
		*WishlistSearch
	Serialize(M) ([]byte, error)
}

func Write[M message[M]](connection net.Conn, message M) (int, error) {
	m, err := message.Serialize(message)
	if err != nil {
		return 0, err
	}

	return internal.MessageWrite(connection, m, false)
}
