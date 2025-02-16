package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const ParentSpeedRatioCode Code = 84

type ParentSpeedRatio struct {
	SpeedRatio int
}

func (p *ParentSpeedRatio) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 84
	if err != nil {
		return err
	}

	if code != uint32(ParentSpeedRatioCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ParentSpeedRatioCode, code))
	}

	p.SpeedRatio, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	return nil
}
