package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const AcceptChildrenCode soul.ServerCode = 100

type AcceptChildren struct{}

func (a AcceptChildren) Serialize(accept bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(AcceptChildrenCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteBool(buf, accept)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
