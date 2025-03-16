package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomUsers Code = 133

type PrivateRoomUsers struct {
	Room  string
	Users []string
}

func (p *PrivateRoomUsers) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 133
	if err != nil {
		return err
	}

	if code != uint32(CodePrivateRoomUsers) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePrivateRoomUsers, code))
	}

	p.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	users, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for range int(users) {
		user, err := internal.ReadString(reader)
		if err != nil {
			return err
		}

		p.Users = append(p.Users, user)
	}

	return nil
}
