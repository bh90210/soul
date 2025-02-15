package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const SendUploadSpeedCode soul.UInt = 121

type SendUploadSpeed struct{}

func (s SendUploadSpeed) Serialize(speed int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, SendUploadSpeedCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(speed))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
