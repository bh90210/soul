package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const UserJoinedRoomCode soul.CodeServer = 16

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
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 16
	if err != nil {
		return err
	}

	if code != uint32(UserJoinedRoomCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", UserJoinedRoomCode, code))
	}

	u.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	u.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	status, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	u.Status = UserStatus(status)

	u.Speed, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	u.Uploads, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	u.Files, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	u.Directories, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	u.Slots, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	u.CountryCode, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
