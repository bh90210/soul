package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const RoomSearchCode soul.ServerCode = 120

type RoomSearch struct{}

func (r RoomSearch) Serialize(room string, token uint32, searchQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(RoomSearchCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, token)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, searchQuery)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
