package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const ParentSpeedRatioCode soul.ServerCode = 84

type ParentSpeedRatio struct {
	SpeedRatio int
}

func (p *ParentSpeedRatio) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 84
	if err != nil {
		return err
	}

	if code != uint32(ParentSpeedRatioCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ParentSpeedRatioCode, code))
	}

	p.SpeedRatio, err = internal.ReadUint32ToInt(reader)
	return err
}
