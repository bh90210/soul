package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const RoomTickerCode soul.UInt = 113

type RoomTicker struct {
	Room  string
	Users []UserTickers
}

type UserTickers struct {
	Username string
	Tickers  string
}

func (r *RoomTicker) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 113
	if err != nil {
		return err
	}

	if code != RoomTickerCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", RoomTickerCode, code))
	}

	users, err := soul.ReadUInt(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(users); i++ {
		var user UserTickers

		user.Username, err = soul.ReadString(reader)
		if err != nil {
			return err
		}

		user.Tickers, err = soul.ReadString(reader)
		if err != nil {
			return err
		}

		r.Users = append(r.Users, user)
	}

	return nil
}
