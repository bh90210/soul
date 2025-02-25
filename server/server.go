// Package server messages are used by clients to interface with the server
// over a connection (TCP).
package server

//go:generate stringer -type Code -trimprefix Code
//go:generate stringer -type UserStatus -trimprefix UserStatus

import (
	"io"
	"net"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// ConnectionType represents the type of server 'S' connection.
const ConnectionType soul.ConnectionType = "S"

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

// Code represents the type of server message.
type Code soul.CodeServer

const (
	CodeLogin Code = 1
	CodeLoginFailed
)

// MessageRead reads a message from a server connection. It reads the size of the message
// and the code of the message. It then reads the message from the connection and
// returns the message, the size of the message, the code of the message and an error.
func MessageRead(connection net.Conn) (io.Reader, int, soul.CodeServer, error) {
	return internal.MessageRead(soul.CodeServer(0), connection)
}

// MessageWrite writes a message to a server connection. It writes the size of the message
// and the code of the message. It then writes the message to the connection and returns
// the number of bytes written and an error.
func MessageWrite(connection net.Conn, message []byte) (int, error) {
	return internal.MessageWrite(connection, message)
}
