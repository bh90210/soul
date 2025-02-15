package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const ParentMinSpeedCode soul.UInt = 83

type ParentMinSpeed struct {
	MinSpeed int
}

func (p *ParentMinSpeed) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 83
	if err != nil {
		return err
	}

	if code != ParentMinSpeedCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ParentMinSpeedCode, code))
	}

	minSpeed, err := soul.ReadUInt(reader)
	if err != nil {
		return err
	}

	p.MinSpeed = int(minSpeed)

	return nil
}
