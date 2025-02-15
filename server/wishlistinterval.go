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
	soul.ReadUInt(reader)         // size
	code := soul.ReadUInt(reader) // code 104
	if code != WishlistIntervalCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", WishlistIntervalCode, code))
	}

	interval := soul.ReadUInt(reader)
	w.Interval = int(interval)

	return nil
}
