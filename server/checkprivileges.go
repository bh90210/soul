package server

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bh90210/soul"
)

const CheckPrivilegesCode Code = 92

type CheckPrivileges struct {
	TimeLeft int
}

func (c CheckPrivileges) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(CheckPrivilegesCode))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (c *CheckPrivileges) Deserialize(reader *bytes.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 92
	if err != nil {
		return err
	}

	if code != uint32(CheckPrivilegesCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CheckPrivilegesCode, code))
	}

	c.TimeLeft, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
