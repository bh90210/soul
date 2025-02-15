package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const UserLeftRoomCode soul.UInt = 17

type UserLeftRoom struct {
	Room     string
	Username string
}

func (u *UserLeftRoom) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 17
	if err != nil {
		return err
	}

	if code != UserLeftRoomCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", UserLeftRoomCode, code))
	}

	u.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	u.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
