package setwaitport

import (
	"bytes"
	"encoding/binary"

	"github.com/bh90210/soul"
)

// Code SetWaitPort.
const Code soul.UInt = 2

// Serialize accepts a port number and returns a serialized byte array.
func Serialize(port soul.UInt) []byte {
	buf := new(bytes.Buffer)
	soul.WriteUInt(buf, Code)

	binary.Write(buf, binary.LittleEndian, port)

	return soul.Pack(buf.Bytes())
}
