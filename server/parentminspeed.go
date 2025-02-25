package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const ParentMinSpeedCode soul.CodeServer = 83

type ParentMinSpeed struct {
	MinSpeed int
}

func (p *ParentMinSpeed) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 83
	if err != nil {
		return err
	}

	if code != uint32(ParentMinSpeedCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ParentMinSpeedCode, code))
	}

	p.MinSpeed, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
