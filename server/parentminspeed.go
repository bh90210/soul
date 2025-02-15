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
	soul.ReadUInt(reader)         // size
	code := soul.ReadUInt(reader) // code 83
	if code != ParentMinSpeedCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ParentMinSpeedCode, code))
	}

	minSpeed := soul.ReadUInt(reader)
	p.MinSpeed = int(minSpeed)

	return nil
}
