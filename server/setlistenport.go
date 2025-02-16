package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

// SetListenPortCode SetWaitPort.
const SetListenPortCode Code = 2

// SetListenPort SetListenPort.
type SetListenPort struct{}

// Serialize accepts a port number and returns a serialized byte array.
func (s SetListenPort) Serialize(port int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(SetListenPortCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(port))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
