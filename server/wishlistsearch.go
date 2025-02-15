package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const WishlistSearchCode soul.UInt = 103

type WishlistSearch struct{}

func (w WishlistSearch) Serialize(token int, serachQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, WishlistSearchCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(token))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, serachQuery)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
