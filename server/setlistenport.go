package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

// SetListenPortCode SetWaitPort.
const SetListenPortCode soul.UInt = 2

// SetListenPort SetListenPort.
type SetListenPort struct{}

// Serialize accepts a port number and returns a serialized byte array.
func (s SetListenPort) Serialize(port soul.UInt) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, SetListenPortCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, port)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
