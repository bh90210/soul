package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const ParentMinSpeedCode Code = 83

type ParentMinSpeed struct {
	MinSpeed int
}

func (p *ParentMinSpeed) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 83
	if err != nil {
		return err
	}

	if code != uint32(ParentMinSpeedCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ParentMinSpeedCode, code))
	}

	p.MinSpeed, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
