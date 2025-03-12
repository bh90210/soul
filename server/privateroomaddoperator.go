package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomAddOperator Code = 143

type PrivateRoomAddOperator struct {
	Room     string
	Username string
}

func (p PrivateRoomAddOperator) Serialize(room, username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodePrivateRoomAddOperator))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (p *PrivateRoomAddOperator) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 143
	if err != nil {
		return err
	}

	if code != uint32(CodePrivateRoomAddOperator) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePrivateRoomAddOperator, code))
	}

	p.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	p.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
