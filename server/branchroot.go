package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodeBranchRoot Code = 127

type BranchRoot struct {
	Root string
}

func (b *BranchRoot) Serialize(message *BranchRoot) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeBranchRoot))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Root)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
