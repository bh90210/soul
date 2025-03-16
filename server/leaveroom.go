package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeLeaveRoom Code = 15

type LeaveRoom struct {
	Room string
}

func (l *LeaveRoom) Serialize(message *LeaveRoom) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeLeaveRoom))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Room)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (l *LeaveRoom) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 15
	if err != nil {
		return err
	}

	if code != uint32(CodeLeaveRoom) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeLeaveRoom, code))
	}

	l.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
