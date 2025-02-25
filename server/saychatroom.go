package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const SayChatroomCode soul.CodeServer = 13

type SayChatroom struct {
	Room     string
	Message  string
	Username string
}

func (s SayChatroom) Serialize(room, message string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(SayChatroomCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (s *SayChatroom) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 13
	if err != nil {
		return err
	}

	if code != uint32(SayChatroomCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", SayChatroomCode, code))
	}

	s.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	s.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	s.Message, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
