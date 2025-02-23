package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const PrivateRoomToggleCode soul.ServerCode = 141

type PrivateRoomToggle struct {
	Enabled bool
}

func (p PrivateRoomToggle) Serialize(enabled bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(PrivateRoomToggleCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteBool(buf, enabled)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (p *PrivateRoomToggle) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 141
	if err != nil {
		return err
	}

	if code != uint32(PrivateRoomToggleCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PrivateRoomToggleCode, code))
	}

	p.Enabled, err = internal.ReadBool(reader)
	if err != nil {
		return err
	}

	return nil
}
