package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const SetStatusCode soul.UInt = 28

type SetStatus struct{}

func (s SetStatus) Serialize(status soul.UserStatusCode) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, SetStatusCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(status))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
