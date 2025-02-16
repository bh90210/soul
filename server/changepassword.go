package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const ChangePasswordCode Code = 142

type ChangePassword struct {
	Pass string
}

func (c ChangePassword) Serialize(pass string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(ChangePasswordCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, pass)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (c *ChangePassword) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 142
	if err != nil {
		return err
	}

	if code != uint32(ChangePasswordCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ChangePasswordCode, code))
	}

	c.Pass, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
