package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodeHaveNoParent Code = 71

type HaveNoParent struct {
	Have bool
}

func (h *HaveNoParent) Serialize(message *HaveNoParent) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeHaveNoParent))
	if err != nil {
		return nil, err
	}

	err = internal.WriteBool(buf, message.Have)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
