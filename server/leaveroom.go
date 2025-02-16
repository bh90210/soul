package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const LeaveRoomCode Code = 15

type LeaveRoom struct {
	Room string
}

func (l LeaveRoom) Serialize(room string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(LeaveRoomCode))
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
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 15
	if err != nil {
		return err
	}

	if code != uint32(LeaveRoomCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", LeaveRoomCode, code))
	}

	l.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
