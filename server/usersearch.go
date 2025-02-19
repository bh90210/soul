package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const UserSearchCode Code = 42

type UserSearch struct{}

func (u UserSearch) Serialize(username string, token uint32, searchQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(UserSearchCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, token)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, searchQuery)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
