package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const BranchLevelCode soul.UInt = 126

type BranchLevel struct{}

func (b BranchLevel) Serialize(level int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, BranchLevelCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(level))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
