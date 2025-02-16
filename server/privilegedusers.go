package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivilegedUsersCode Code = 69

type PrivilegedUsers struct {
	Users []string
}

func (p *PrivilegedUsers) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 69
	if err != nil {
		return err
	}

	if code != uint32(PrivilegedUsersCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivilegedUsersCode, code))
	}

	numberOfUsers, err := soul.ReadUint32(reader)
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
