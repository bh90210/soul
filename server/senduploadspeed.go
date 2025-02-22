package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const SendUploadSpeedCode soul.ServerCode = 121

type SendUploadSpeed struct{}

func (s SendUploadSpeed) Serialize(speed int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(SendUploadSpeedCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(speed))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
