package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const UnwatchUserCode soul.ServerCode = 6

type UnwatchUser struct{}

func (u UnwatchUser) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(UnwatchUserCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
