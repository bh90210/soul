package server

import (
	"fmt"
	"io"

	"errors"

	"github.com/bh90210/soul"
)

const WishlistIntervalCode Code = 104

type WishlistInterval struct {
	Interval int
}

func (w *WishlistInterval) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 104
	if err != nil {
		return err
	}

	if code != uint32(WishlistIntervalCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", WishlistIntervalCode, code))
	}

	w.Interval, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
