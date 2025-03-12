package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomOperatorRemoved Code = 146

type PrivateRoomOperatorRemoved struct {
	Room string
}

func (p *PrivateRoomOperatorRemoved) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // Size.
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	if code != uint32(CodePrivateRoomOperatorRemoved) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePrivateRoomOperatorRemoved, code))
	}

	p.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
