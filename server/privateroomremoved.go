package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomRemoved Code = 140

type PrivateRoomRemoved struct {
	Room string
}

func (p *PrivateRoomRemoved) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 139
	if err != nil {
		return err
	}

	if code != uint32(CodePrivateRoomRemoved) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePrivateRoomRemoved, code))
	}

	p.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
