package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeAdminMessage Code = 66

type AdminMessage struct {
	Message string
}

func (a *AdminMessage) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 66
	if err != nil {
		return err
	}

	if code != uint32(CodeAdminMessage) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeAdminMessage, code))
	}

	a.Message, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
