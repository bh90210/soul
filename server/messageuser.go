package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const MessageUserCode soul.CodeServer = 22

type MessageUser struct {
	UserID    int
	Timestamp int
	Username  string
	Message   string
	New       bool
}

func (m MessageUser) Serialize(username, message string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(MessageUserCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (m *MessageUser) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 22
	if err != nil {
		return err
	}

	if code != uint32(MessageUserCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", MessageUserCode, code))
	}

	m.UserID, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	m.Timestamp, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	m.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	m.Message, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	m.New, err = internal.ReadBool(reader)
	if err != nil {
		return err
	}

	return nil
}
