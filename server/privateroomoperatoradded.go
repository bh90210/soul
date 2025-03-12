package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomOperatorAdded Code = 145

type PrivateRoomOperatorAdded struct {
	Room string
}

func (p *PrivateRoomOperatorAdded) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 145
	if err != nil {
		return err
	}

	if code != uint32(CodePrivateRoomOperatorAdded) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePrivateRoomOperatorAdded, code))
	}

	p.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
