// Package distributed messages are sent to peers over a D connection (TCP),
// and are used for the distributed search network.
// Only a single active connection to a peer is allowed.
package distributed

//go:generate stringer -type=Code -trimprefix=Code

import (
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
	"github.com/bh90210/soul/peer"
)

// ConnectionType represents the type of distributed 'D' connection.
const ConnectionType soul.ConnectionType = "D"

// Code represents the type of distributed message.
type Code int

// Read reads a message from a distributed connection. It reads the size of the message
// and the code of the message. It then reads the message from the connection and
// returns the message, the size of the message, the code of the message and an error.
func Read(connection io.Reader) (io.Reader, int, Code, error) {
	r, s, c, err := internal.MessageRead(internal.CodeDistributed(0), connection, false)
	return r, int(s), Code(c), err
}

// message writes a message to a distributed connection. It writes the size of the message
// and the code of the message. It then writes the message to the connection and returns
// the number of bytes written and an error.
type message[M any] interface {
	*BranchLevel |
		*BranchRoot |
		*EmbeddedMessage |
		*Search |
		*peer.PeerInit
	Serialize(M) ([]byte, error)
}

// Write writes a message to a distributed connection. It writes the size of the message
// and the code of the message. It then writes the message to the connection and returns
// the number of bytes written and an error.
func Write[M message[M]](connection io.Writer, message M) (int, error) {
	m, err := message.Serialize(message)
	if err != nil {
		return 0, err
	}

	return internal.MessageWrite(connection, m, false)
}
