package server

import (
	"fmt"
	"io"

	"errors"

	"github.com/bh90210/soul"
)

const WishlistIntervalCode soul.UInt = 104

type WishlistInterval struct {
	Interval int
}

func (w *WishlistInterval) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 104
	if err != nil {
		return err
	}

	if code != WishlistIntervalCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", WishlistIntervalCode, code))
	}

	interval, err := soul.ReadUInt(reader)
	if err != nil {
		return err
	}

	w.Interval = int(interval)

	return nil
}
