package server

import (
	"bytes"
	"encoding/binary"

	"github.com/bh90210/soul"
)

// SetWaitPortCode SetWaitPort.
const SetWaitPortCode soul.UInt = 2

// SetWaitPort SetWaitPort.
type SetWaitPort struct{}

// Serialize accepts a port number and returns a serialized byte array.
func (s SetWaitPort) Serialize(port soul.UInt) ([]byte, error) {
	buf := new(bytes.Buffer)
	soul.WriteUInt(buf, SetWaitPortCode)

	err := binary.Write(buf, binary.LittleEndian, port)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes()), nil
}
