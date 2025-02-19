// Package server messages are used by clients to interface with the server
// over a connection (TCP).
package server

import (
	"bytes"
	"errors"
	"io"
	"net"

	"github.com/bh90210/soul"
)

type Code int

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

var ErrDifferentPacketSize = errors.New("the declared size of the package does not match the size of the actual read")

func ReadMessage(connection net.Conn) (io.Reader, int, Code, error) {
	message := new(bytes.Buffer)

	// We need to make two reads from the connection to determine the code of the message.
	// Because we need these information down the line we TeeRead them to the message.
	// Note that there is no "message header" in the protocol, we just read the size and code
	// from the "head" of the packet.
	messageHeader := io.TeeReader(connection, message)

	// Read the size of the packet.
	size, err := soul.ReadUint32(messageHeader)
	if err != nil {
		return nil, 0, 0, err
	}

	// Read the code of the message.
	code, err := soul.ReadUint32(messageHeader)
	if err != nil {
		return nil, 0, 0, err
	}

	// Now we simply copy a packet size read from the connection to the message buffer.
	// This continues writing the message buffer from where the TeeReader left off.
	// The size of the actual message read needs -4 to account for the packet
	// size and code reads that happened above.
	n, err := io.CopyN(message, connection, int64(size-4))
	if err != nil {
		return nil, 0, 0, err
	}

	// Conversely, we need to add 4 to the size of the total read to account for the
	// size and code reads that are missing from CopyN.
	n += 4

	if int64(size) != n {
		return nil, 0, 0, ErrDifferentPacketSize
	}

	return message, int(size), Code(code), nil
}

func WriteMessage(connection net.Conn, message []byte) (int, error) {
	n, err := connection.Write(message)
	if err != nil {
		return 0, err
	}

	return n, nil
}
