package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

// CodeSetListenPort SetWaitPort.
const CodeSetListenPort Code = 2

// SetListenPort SetListenPort.
type SetListenPort struct{}

// Serialize accepts a port number and returns a serialized byte array.
func (s SetListenPort) Serialize(port uint32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeSetListenPort))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, port)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
