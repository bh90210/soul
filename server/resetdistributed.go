package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const ResetDistributedCode soul.ServerCode = 130

type ResetDistributed struct{}

func (r *ResetDistributed) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 130
	if err != nil {
		return err
	}

	if code != uint32(ResetDistributedCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ResetDistributedCode, code))
	}

	return nil
}
