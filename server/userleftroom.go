package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeUserLeftRoom Code = 17

type UserLeftRoom struct {
	Room     string
	Username string
}

func (u *UserLeftRoom) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 17
	if err != nil {
		return err
	}

	if code != uint32(CodeUserLeftRoom) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeUserLeftRoom, code))
	}

	u.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	u.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
