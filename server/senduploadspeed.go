package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodeSendUploadSpeed Code = 121

type SendUploadSpeed struct {
	Speed int
}

func (s *SendUploadSpeed) Serialize(message *SendUploadSpeed) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeSendUploadSpeed))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Speed))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
