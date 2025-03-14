package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeRoomTicker Code = 113

type RoomTicker struct {
	Room  string
	Users []UserTickers
}

type UserTickers struct {
	Username string
	Tickers  string
}

func (r *RoomTicker) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 113
	if err != nil {
		return err
	}

	if code != uint32(CodeRoomTicker) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeRoomTicker, code))
	}

	users, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(users); i++ {
		var user UserTickers

		user.Username, err = internal.ReadString(reader)
		if err != nil {
			return err
		}

		user.Tickers, err = internal.ReadString(reader)
		if err != nil {
			return err
		}

		r.Users = append(r.Users, user)
	}

	return nil
}
