// Package distributed messages are sent to peers over a D connection (TCP),
// and are used for the distributed search network.
// Only a single active connection to a peer is allowed.
package distributed

import (
	"io"
	"net"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

func MessageRead(connection net.Conn) (io.Reader, int, soul.DistributedCode, error) {
	return internal.MessageRead(soul.DistributedCode(0), connection)
}

func MessageWrite(connection net.Conn, message []byte) (int, error) {
	return internal.MessageWrite(connection, message)
}
