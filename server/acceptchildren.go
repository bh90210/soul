package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodeAcceptChildren Code = 100

type AcceptChildren struct{}

func (AcceptChildren) Serialize(accept bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeAcceptChildren))
	if err != nil {
		return nil, err
	}

	err = internal.WriteBool(buf, accept)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
