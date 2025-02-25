package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const SendUploadSpeedCode soul.CodeServer = 121

type SendUploadSpeed struct{}

func (s SendUploadSpeed) Serialize(speed int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(SendUploadSpeedCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(speed))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
