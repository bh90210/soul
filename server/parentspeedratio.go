package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const ParentSpeedRatioCode soul.UInt = 84

type ParentSpeedRatio struct {
	SpeedRatio int
}

func (p *ParentSpeedRatio) Deserialize(reader io.Reader) error {
	soul.ReadUInt(reader)         // size
	code := soul.ReadUInt(reader) // code 84
	if code != ParentSpeedRatioCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ParentSpeedRatioCode, code))
	}

	speedRatio := soul.ReadUInt(reader)
	p.SpeedRatio = int(speedRatio)

	return nil
}
