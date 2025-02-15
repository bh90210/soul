package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivateRoomAddUserCode soul.UInt = 134

type PrivateRoomAddUser struct {
	Room     string
	Username string
}

func (p PrivateRoomAddUser) Serialize(room, username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, PrivateRoomAddUserCode)
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

func (p *PrivateRoomAddUser) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 134
	if err != nil {
		return err
	}

	if code != PrivateRoomAddUserCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomAddUserCode, code))
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
