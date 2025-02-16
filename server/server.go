package server

import (
	"bytes"
	"errors"
	"io"
	"net"

	"github.com/bh90210/soul"
)

type Code int

var ErrDifferentPacketSize = errors.New("different packet size")

func ReadMessage(connection net.Conn) (io.Reader, int, Code, error) {
	message := new(bytes.Buffer)

	// We need to make two reads from the connection to determine the code of the message.
	// Because we need those information down the line we TeeRead them for later use.
	messageHeader := io.TeeReader(connection, message)

	// Read the size of the packet.
	packetSize, err := soul.ReadUint32(messageHeader)
	if err != nil {
		return nil, 0, 0, err
	}

	// Read the code of the message.
	code, err := soul.ReadUint32(messageHeader)
	if err != nil {
		return nil, 0, 0, err
	}

	// Now we simply copy the a packet size read from the connection to the message buffer.
	// The size of the actual message needs -4 to account for the packet size and code reads.
	n, err := io.CopyN(message, connection, int64(packetSize-4))
	if err != nil {
		return nil, 0, 0, err
	}

	if int64(packetSize) != n+4 {
		return nil, 0, 0, ErrDifferentPacketSize
	}

	return message, int(packetSize), Code(code), nil
}

func SendMessage(connection net.Conn, message []byte) (int, error) {
	n, err := connection.Write(message)
	if err != nil {
		return 0, err
	}

	return n, nil
}
