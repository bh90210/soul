package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const PrivilegedUsersCode soul.CodeServer = 69

type PrivilegedUsers struct {
	Users []string
}

func (p *PrivilegedUsers) Deserialize(reader io.Reader) (err error) {
	_, err = internal.ReadUint32(reader) // size
	if err != nil {
		return
	}

	code, err := internal.ReadUint32(reader) // code 69
	if err != nil {
		return
	}

	if code != uint32(PrivilegedUsersCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivilegedUsersCode, code))
	}

	numberOfUsers, err := internal.ReadUint32(reader)
	if err != nil {
		return
	}

	for i := 0; i < int(numberOfUsers); i++ {
		var user string
		user, err = internal.ReadString(reader)
		if err != nil && !errors.Is(err, io.EOF) {
			return
		}

		p.Users = append(p.Users, user)
	}

	return
}
