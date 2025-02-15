package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const ResetDistributedCode soul.UInt = 130

type ResetDistributed struct{}

func (r *ResetDistributed) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 130
	if err != nil {
		return err
	}

	if code != ResetDistributedCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ResetDistributedCode, code))
	}

	return nil
}
