package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const RoomTickerAddCode Code = 114

type RoomTickerAdd struct {
	Room     string
	Username string
	Ticker   string
}

func (r *RoomTickerAdd) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 114
	if err != nil {
		return err
	}

	if code != uint32(RoomTickerAddCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", RoomTickerAddCode, code))
	}

	r.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	r.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	r.Ticker, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
