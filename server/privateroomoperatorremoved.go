package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivateRoomOperatorRemovedCode soul.UInt = 146

type PrivateRoomOperatorRemoved struct {
	Room string
}

func (p *PrivateRoomOperatorRemoved) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 146
	if err != nil {
		return err
	}

	if code != PrivateRoomOperatorRemovedCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomOperatorRemovedCode, code))
	}

	p.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
