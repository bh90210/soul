package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivateRoomOperatorRemovedCode soul.ServerCode = 146

type PrivateRoomOperatorRemoved struct {
	Room string
}

func (p *PrivateRoomOperatorRemoved) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 146
	if err != nil {
		return err
	}

	if code != uint32(PrivateRoomOperatorRemovedCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomOperatorRemovedCode, code))
	}

	p.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
