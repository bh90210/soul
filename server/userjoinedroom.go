package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const UserJoinedRoomCode soul.ServerCode = 16

type UserJoinedRoom struct {
	Room        string
	Username    string
	Status      UserStatus
	Speed       int
	Uploads     int
	Files       int
	Directories int
	Slots       int
	CountryCode string
}

func (u *UserJoinedRoom) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 16
	if err != nil {
		return err
	}

	if code != uint32(UserJoinedRoomCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", UserJoinedRoomCode, code))
	}

	u.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	u.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	status, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	u.Status = UserStatus(status)

	u.Speed, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	u.Uploads, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	u.Files, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	u.Directories, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	u.Slots, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	u.CountryCode, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
