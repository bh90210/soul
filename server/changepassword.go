package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const ChangePasswordCode soul.CodeServer = 142

type ChangePassword struct {
	Pass string
}

func (c ChangePassword) Serialize(pass string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(ChangePasswordCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, pass)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (c *ChangePassword) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 142
	if err != nil {
		return err
	}

	if code != uint32(ChangePasswordCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ChangePasswordCode, code))
	}

	c.Pass, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
