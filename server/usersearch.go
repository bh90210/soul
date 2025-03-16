package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeUserSearch Code = 42

type UserSearch struct {
	Username    string
	Token       soul.Token
	SearchQuery string
}

func (u *UserSearch) Serialize(message *UserSearch) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeUserSearch))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.SearchQuery)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
