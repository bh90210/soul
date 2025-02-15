package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const SayChatroomCode soul.UInt = 13

type SayChatroom struct {
	Room     string
	Message  string
	Username string
}

func (s SayChatroom) Serialize(room, message string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, SayChatroomCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, message)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (s *SayChatroom) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 13
	if err != nil {
		return err
	}

	if code != SayChatroomCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", SayChatroomCode, code))
	}

	s.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	s.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	s.Message, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
