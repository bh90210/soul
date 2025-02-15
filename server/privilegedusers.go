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
	soul.ReadUInt(reader)         // size
	code := soul.ReadUInt(reader) // code 69
	if code != PrivilegedUsersCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivilegedUsersCode, code))
	}

	numberOfUsers := soul.ReadUInt(reader)

	for i := 0; i < int(numberOfUsers); i++ {
		user := soul.ReadString(reader)

		p.Users = append(p.Users, user)
	}

	return nil
}
