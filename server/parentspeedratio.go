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
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 84
	if err != nil {
		return err
	}

	if code != ParentSpeedRatioCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ParentSpeedRatioCode, code))
	}

	speedRatio, err := soul.ReadUInt(reader)
	if err != nil {
		return err
	}

	p.SpeedRatio = int(speedRatio)

	return nil
}
