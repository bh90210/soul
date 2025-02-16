package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivateRoomOperatorAddedCode Code = 145

type PrivateRoomOperatorAdded struct {
	Room string
}

func (p *PrivateRoomOperatorAdded) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 145
	if err != nil {
		return err
	}

	if code != uint32(PrivateRoomOperatorAddedCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomOperatorAddedCode, code))
	}

	p.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
