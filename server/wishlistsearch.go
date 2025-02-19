package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const WishlistSearchCode Code = 103

type WishlistSearch struct{}

func (w WishlistSearch) Serialize(token uint32, serachQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(WishlistSearchCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, token)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, serachQuery)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
