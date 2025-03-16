package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

// CodeSetListenPort SetWaitPort.
const CodeSetListenPort Code = 2

// SetListenPort SetListenPort.
type SetListenPort struct {
	Port           int
	ObfuscatedPort int
}

// Serialize accepts a port number and returns a serialized byte array.
func (s SetListenPort) Serialize(message *SetListenPort) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeSetListenPort))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Port))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.ObfuscatedPort))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
