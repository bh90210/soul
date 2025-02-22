package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const RoomSearchCode soul.ServerCode = 120

type RoomSearch struct{}

func (r RoomSearch) Serialize(room string, token uint32, searchQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(RoomSearchCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, token)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, searchQuery)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
