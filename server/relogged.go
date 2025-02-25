package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const ReloggedCode soul.CodeServer = 41

type Relogged struct{}

func (r *Relogged) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 41
	if err != nil {
		return err
	}

	if code != uint32(ReloggedCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ReloggedCode, code))
	}

	return nil
}
