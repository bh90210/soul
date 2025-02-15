package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivateRoomAddedCode soul.UInt = 139

type PrivateRoomAdded struct {
	Room string
}

func (p *PrivateRoomAdded) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 139
	if err != nil {
		return err
	}

	if code != PrivateRoomAddedCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomAddedCode, code))
	}

	p.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
