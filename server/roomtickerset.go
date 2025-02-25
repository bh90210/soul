package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const RoomTickerSetCode soul.CodeServer = 116

type RoomTickerSet struct{}

func (r RoomTickerSet) Serialize(room, ticker string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(RoomTickerSetCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, ticker)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
