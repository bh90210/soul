package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const UserLeftRoomCode Code = 17

type UserLeftRoom struct {
	Room     string
	Username string
}

func (u *UserLeftRoom) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 17
	if err != nil {
		return err
	}

	if code != uint32(UserLeftRoomCode) {
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
