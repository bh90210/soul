// Package distributed messages are sent to peers over a D connection (TCP),
// and are used for the distributed search network.
// Only a single active connection to a peer is allowed.
package distributed

//go:generate stringer -type=Code -trimprefix=Code

import (
	"io"
	"net"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// ConnectionType represents the type of distributed 'D' connection.
const ConnectionType soul.ConnectionType = "D"

// Code represents the type of distributed message.
type Code soul.CodeDistributed

const (
	CodeSearch          Code = 3
	CodeBranchLevel     Code = 4
	CodeBranchRoot      Code = 5
	CodeEmbeddedMessage Code = 93
)

// MessageRead reads a message from a distributed connection. It reads the size of the message
// and the code of the message. It then reads the message from the connection and
// returns the message, the size of the message, the code of the message and an error.
func MessageRead(connection net.Conn) (io.Reader, int, soul.CodeDistributed, error) {
	return internal.MessageRead(soul.CodeDistributed(0), connection)
}

// MessageWrite writes a message to a distributed connection. It writes the size of the message
// and the code of the message. It then writes the message to the connection and returns
// the number of bytes written and an error.
func MessageWrite(connection net.Conn, message []byte) (int, error) {
	return internal.MessageWrite(connection, message)
}
