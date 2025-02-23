package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const PrivateRoomDisownCode soul.ServerCode = 137

type PrivateRoomDisown struct{}

func (p PrivateRoomDisown) Serialize(room string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(PrivateRoomDisownCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
