package server

import (
	"bytes"
	"io"
	"log"

	"github.com/bh90210/soul"
)

func Message(connection io.Reader) (io.Reader, soul.UInt, soul.UInt) {
	messageCopy := new(bytes.Buffer)
	message := io.TeeReader(connection, messageCopy)

	// Read the messageSize of the message.
	packetSize := soul.ReadUInt(message)

	// Read the code.
	code := soul.ReadUInt(message)

	// The size of the actual message needs -4 to account for the packetSize and code.
	messageSize := packetSize - 4

	sizeSoFar := 0
	for {
		p := make([]byte, int(messageSize)-sizeSoFar)
		n, err := message.Read(p)
		if err != nil {
			log.Fatal("read error", err)
		}

		sizeSoFar += n

		if sizeSoFar == int(messageSize) {
			break
		}
	}

	return messageCopy, packetSize, code
}
