package server

import (
	"fmt"
	"io"

	"errors"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeWishlistInterval Code = 104

type WishlistInterval struct {
	Interval int
}

func (w *WishlistInterval) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 104
	if err != nil {
		return err
	}

	if code != uint32(CodeWishlistInterval) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeWishlistInterval, code))
	}

	w.Interval, err = internal.ReadUint32ToInt(reader)
	return err
}
