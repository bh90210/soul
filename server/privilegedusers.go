package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivilegedUsersCode soul.UInt = 69

type PrivilegedUsers struct {
	Users []string
}

func (p *PrivilegedUsers) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 69
	if err != nil {
		return err
	}

	if code != PrivilegedUsersCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivilegedUsersCode, code))
	}

	numberOfUsers, err := soul.ReadUInt(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(numberOfUsers); i++ {
		user, err := soul.ReadString(reader)
		if err != nil {
			return err
		}

		p.Users = append(p.Users, user)
	}

	return nil
}
