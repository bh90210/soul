package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const RoomTickerAddCode soul.CodeServer = 114

type RoomTickerAdd struct {
	Room     string
	Username string
	Ticker   string
}

func (r *RoomTickerAdd) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 114
	if err != nil {
		return err
	}

	if code != uint32(RoomTickerAddCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", RoomTickerAddCode, code))
	}

	r.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	r.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	r.Ticker, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
