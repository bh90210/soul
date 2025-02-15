package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const LeaveRoomCode soul.UInt = 15

type LeaveRoom struct {
	Room string
}

func (l LeaveRoom) Serialize(room string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, LeaveRoomCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (l *LeaveRoom) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 15
	if err != nil {
		return err
	}

	if code != LeaveRoomCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", LeaveRoomCode, code))
	}

	l.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
