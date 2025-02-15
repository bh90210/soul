package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const CantCreateRoomCode soul.UInt = 1003

type CantCreateRoom struct {
	Room string
}

func (c *CantCreateRoom) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 1003
	if err != nil {
		return err
	}

	if code != CantCreateRoomCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CantCreateRoomCode, code))
	}

	c.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
