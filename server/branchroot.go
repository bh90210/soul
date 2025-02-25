package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const BranchRootCode soul.CodeServer = 127

type BranchRoot struct{}

func (BranchRoot) Serialize(root string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(BranchRootCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, root)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
