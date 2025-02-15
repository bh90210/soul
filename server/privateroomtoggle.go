package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const PrivateRoomToggleCode soul.UInt = 141

type PrivateRoomToggle struct {
	Enabled bool
}

func (p PrivateRoomToggle) Serialize(enabled bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, PrivateRoomToggleCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteBool(buf, enabled)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (p *PrivateRoomToggle) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 141
	if err != nil {
		return err
	}

	if code != PrivateRoomToggleCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomToggleCode, code))
	}

	p.Enabled, err = soul.ReadBool(reader)
	if err != nil {
		return err
	}

	return nil
}
