package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivateRoomUsersCode soul.UInt = 133

type PrivateRoomUsers struct {
	Room  string
	Users []string
}

func (p *PrivateRoomUsers) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 133
	if err != nil {
		return err
	}

	if code != PrivateRoomUsersCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomUsersCode, code))
	}

	p.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	users, err := soul.ReadUInt(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(users); i++ {
		user, err := soul.ReadString(reader)
		if err != nil {
			return err
		}

		p.Users = append(p.Users, user)
	}

	return nil
}
