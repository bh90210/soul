package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodePing Code = 32

type Ping struct{}

func (p Ping) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodePing))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
