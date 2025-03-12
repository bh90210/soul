package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeRoomTickerRemove Code = 115

type RoomTickerRemove struct {
	Room     string
	Username string
}

func (r *RoomTickerRemove) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 115
	if err != nil {
		return err
	}

	if code != uint32(CodeRoomTickerRemove) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeRoomTickerRemove, code))
	}

	r.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	r.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
