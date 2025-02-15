package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const RoomTickerSetCode soul.UInt = 116

type RoomTickerSet struct{}

func (r RoomTickerSet) Serialize(room, ticker string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, RoomTickerSetCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, ticker)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
