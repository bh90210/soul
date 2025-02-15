package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const UserSearchCode soul.UInt = 42

type UserSearch struct{}

func (u UserSearch) Serialize(username string, token int, searchQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, UserSearchCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(token))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, searchQuery)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
