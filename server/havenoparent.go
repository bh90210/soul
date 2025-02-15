package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const HaveNoParentCode soul.UInt = 71

type HaveNoParent struct{}

func (h HaveNoParent) Serialize(haveParents bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, HaveNoParentCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteBool(buf, haveParents)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
