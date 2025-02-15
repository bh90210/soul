package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const RoomSearchCode soul.UInt = 120

type RoomSearch struct{}

func (r RoomSearch) Serialize(room string, token int, searchQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, RoomSearchCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(token))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, searchQuery)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
