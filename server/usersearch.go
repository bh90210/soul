package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeUserSearch Code = 42

type UserSearch struct{}

func (u UserSearch) Serialize(username string, token soul.Token, searchQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeUserSearch))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, searchQuery)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
