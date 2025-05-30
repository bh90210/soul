package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomRemoveUser Code = 135

type PrivateRoomRemoveUser struct {
	Room     string
	Username string
}

func (p *PrivateRoomRemoveUser) Serialize(message *PrivateRoomRemoveUser) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodePrivateRoomRemoveUser))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Room)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Username)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (p *PrivateRoomRemoveUser) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 135
	if err != nil {
		return err
	}

	if code != uint32(CodePrivateRoomRemoveUser) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePrivateRoomRemoveUser, code))
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
