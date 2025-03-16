package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodeSetStatus Code = 28

type SetStatus struct {
	Status UserStatus
}

func (s *SetStatus) Serialize(message *SetStatus) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeSetStatus))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Status))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
