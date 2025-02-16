package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const RoomTickerRemoveCode Code = 115

type RoomTickerRemove struct {
	Room     string
	Username string
}

func (r *RoomTickerRemove) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 115
	if err != nil {
		return err
	}

	if code != uint32(RoomTickerRemoveCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", RoomTickerRemoveCode, code))
	}

	r.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	r.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
