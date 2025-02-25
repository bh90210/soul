package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CheckPrivilegesCode soul.CodeServer = 92

type CheckPrivileges struct {
	TimeLeft int
}

func (c CheckPrivileges) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CheckPrivilegesCode))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (c *CheckPrivileges) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 92
	if err != nil {
		return err
	}

	if code != uint32(CheckPrivilegesCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CheckPrivilegesCode, code))
	}

	c.TimeLeft, err = internal.ReadUint32ToInt(reader)
	return err
}
