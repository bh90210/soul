package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeChangePassword Code = 142

type ChangePassword struct {
	Pass string
}

func (ChangePassword) Serialize(pass string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeChangePassword))
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

	if code != uint32(CodeChangePassword) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeChangePassword, code))
	}

	c.Pass, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
