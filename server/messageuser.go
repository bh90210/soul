package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const MessageUserCode Code = 22

type MessageUser struct {
	UserID    int
	Timestamp int
	Username  string
	Message   string
	New       bool
}

func (m MessageUser) Serialize(username, message string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(MessageUserCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, message)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (m *MessageUser) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 22
	if err != nil {
		return err
	}

	if code != uint32(MessageUserCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", MessageUserCode, code))
	}

	m.UserID, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	m.Timestamp, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	m.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	m.Message, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	m.New, err = soul.ReadBool(reader)
	if err != nil {
		return err
	}

	return nil
}
