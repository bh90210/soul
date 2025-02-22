package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const BranchLevelCode soul.ServerCode = 126

type BranchLevel struct{}

func (b BranchLevel) Serialize(level int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(BranchLevelCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(level))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
