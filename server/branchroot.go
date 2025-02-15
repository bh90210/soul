package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const BranchRootCode soul.UInt = 127

type BranchRoot struct{}

func (b BranchRoot) Serialize(root string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, BranchRootCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, root)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
