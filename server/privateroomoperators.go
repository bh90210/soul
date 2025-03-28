package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomOperators Code = 148

type PrivateRoomOperators struct {
	Room      string
	Operators []string
}

func (p *PrivateRoomOperators) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 144
	if err != nil {
		return err
	}

	if code != uint32(CodePrivateRoomOperators) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePrivateRoomOperators, code))
	}

	p.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	operators, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for range int(operators) {
		operator, err := internal.ReadString(reader)
		if err != nil {
			return err
		}

		p.Operators = append(p.Operators, operator)
	}

	return nil
}
