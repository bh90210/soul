package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const WishlistSearchCode soul.ServerCode = 103

type WishlistSearch struct{}

func (w WishlistSearch) Serialize(token soul.Token, serachQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(WishlistSearchCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, token.Uint32())
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, serachQuery)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
