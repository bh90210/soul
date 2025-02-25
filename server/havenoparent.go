package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const HaveNoParentCode soul.CodeServer = 71

type HaveNoParent struct{}

func (h HaveNoParent) Serialize(haveParents bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(HaveNoParentCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteBool(buf, haveParents)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
