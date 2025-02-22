package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const AdminMessageCode soul.ServerCode = 66

type AdminMessage struct {
	Message string
}

func (a *AdminMessage) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 66
	if err != nil {
		return err
	}

	if code != uint32(AdminMessageCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", AdminMessageCode, code))
	}

	a.Message, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
