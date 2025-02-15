package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivateRoomOperatorsCode soul.UInt = 144

type PrivateRoomOperators struct {
	Room      string
	Operators []string
}

func (p *PrivateRoomOperators) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 144
	if err != nil {
		return err
	}

	if code != PrivateRoomOperatorsCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomOperatorsCode, code))
	}

	p.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	operators, err := soul.ReadUInt(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(operators); i++ {
		operator, err := soul.ReadString(reader)
		if err != nil {
			return err
		}

		p.Operators = append(p.Operators, operator)
	}

	return nil
}
