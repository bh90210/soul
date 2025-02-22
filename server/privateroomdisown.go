package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const PrivateRoomDisownCode soul.ServerCode = 137

type PrivateRoomDisown struct{}

func (p PrivateRoomDisown) Serialize(room string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(PrivateRoomDisownCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
