package server

import (
	"bytes"
	"encoding/binary"

	"github.com/bh90210/soul"
)

const UnwatchUserCode soul.UInt = 6

type UnwatchUser struct{}

func (u UnwatchUser) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, UnwatchUserCode)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, username)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
