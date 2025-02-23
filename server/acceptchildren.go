package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const AcceptChildrenCode soul.ServerCode = 100

type AcceptChildren struct{}

func (a AcceptChildren) Serialize(accept bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(AcceptChildrenCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteBool(buf, accept)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
