// Package server messages are used by clients to interface with the server
// over a connection (TCP).
package server

import (
	"io"
	"net"

	"github.com/bh90210/soul"
)

// UserStatus represents the status of a user.
type UserStatus int

const (
	// Offline user status.
	Offline UserStatus = iota
	// Away user status.
	Away
	// Online user status.
	Online
)

func MessageRead(connection net.Conn) (io.Reader, int, soul.ServerCode, error) {
	return soul.MessageRead(soul.ServerCode(0), connection)
}

func MessageWrite(connection net.Conn, message []byte) (int, error) {
	return soul.MessageWrite(connection, message)
}
