package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomAdded Code = 139

type PrivateRoomAdded struct {
	Room string
}

func (p *PrivateRoomAdded) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 139
	if err != nil {
		return err
	}

	if code != uint32(CodePrivateRoomAdded) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePrivateRoomAdded, code))
	}

	p.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
