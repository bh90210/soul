package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const SetStatusCode Code = 28

type SetStatus struct{}

func (s SetStatus) Serialize(status UserStatus) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(SetStatusCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(status))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
