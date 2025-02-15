package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const UserJoinedRoomCode soul.UInt = 16

type UserJoinedRoom struct {
	Room        string
	Username    string
	Status      soul.UserStatusCode
	Speed       int
	Uploads     int
	Files       int
	Directories int
	Slots       int
	CountryCode string
}

func (u *UserJoinedRoom) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 16
	if err != nil {
		return err
	}

	if code != UserJoinedRoomCode {
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

	status, err := soul.ReadUInt(reader)
	if err != nil {
		return err
	}

	u.Status = soul.UserStatusCode(status)

	u.Speed, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	u.Uploads, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	u.Files, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	u.Directories, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	u.Slots, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	u.CountryCode, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
