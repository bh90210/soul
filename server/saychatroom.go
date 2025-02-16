package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const SayChatroomCode Code = 13

type SayChatroom struct {
	Room     string
	Message  string
	Username string
}

func (s SayChatroom) Serialize(room, message string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(SayChatroomCode))
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
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 13
	if err != nil {
		return err
	}

	if code != uint32(SayChatroomCode) {
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
