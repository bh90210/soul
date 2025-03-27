// Package file messages are sent to peers over a F connection (TCP),
// and do not have messages codes associated with them.
package file

import (
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// ConnectionType represents the type of file 'F' connection.
const ConnectionType soul.ConnectionType = "F"

type message[M any] interface {
	*TransferInit | *Offset
	Serialize(M) ([]byte, error)
}

func Write[M message[M]](connection io.Writer, message M) (int, error) {
	m, err := message.Serialize(message)
	if err != nil {
		return 0, err
	}

	return internal.MessageWrite(connection, m, false)
}
