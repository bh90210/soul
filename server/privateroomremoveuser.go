package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivateRoomRemoveUserCode soul.UInt = 135

type PrivateRoomRemoveUser struct {
	Room     string
	Username string
}

func (p PrivateRoomRemoveUser) Serialize(room, username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, PrivateRoomRemoveUserCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (p *PrivateRoomRemoveUser) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 135
	if err != nil {
		return err
	}

	if code != PrivateRoomRemoveUserCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomRemoveUserCode, code))
	}

	p.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	p.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
