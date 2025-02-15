package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivateRoomOperatorAddedCode soul.UInt = 145

type PrivateRoomOperatorAdded struct {
	Room string
}

func (p *PrivateRoomOperatorAdded) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 145
	if err != nil {
		return err
	}

	if code != PrivateRoomOperatorAddedCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomOperatorAddedCode, code))
	}

	p.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
