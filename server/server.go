package server

import (
	"bytes"
	"io"
	"net"

	"github.com/bh90210/soul"
)

func ReadMessage(connection net.Conn) (io.Reader, soul.UInt, soul.UInt, error) {
	messageCopy := new(bytes.Buffer)
	message := io.TeeReader(connection, messageCopy)

	// Read the messageSize of the message.
	packetSize, err := soul.ReadUInt(message)
	if err != nil {
		return nil, 0, 0, err
	}

	// Read the code.
	code, err := soul.ReadUInt(message)
	if err != nil {
		return nil, 0, 0, err
	}

	// The size of the actual message needs -4 to account for the packetSize and code.
	messageSize := packetSize - 4

	sizeSoFar := 0
	for {
		p := make([]byte, int(messageSize)-sizeSoFar)
		n, err := message.Read(p)
		if err != nil {
			return nil, 0, 0, err
		}

		sizeSoFar += n

		if sizeSoFar == int(messageSize) {
			break
		}
	}

	return messageCopy, packetSize, code, nil
}

func SendMessage(connection net.Conn, message []byte) (int, error) {
	n, err := connection.Write(message)
	if err != nil {
		return 0, err
	}

	return n, nil
}
