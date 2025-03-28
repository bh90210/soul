package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeCantCreateRoom Code = 1003

type CantCreateRoom struct {
	Room string
}

func (c *CantCreateRoom) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 1003
	if err != nil {
		return err
	}

	if code != uint32(CodeCantCreateRoom) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeCantCreateRoom, code))
	}

	c.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
