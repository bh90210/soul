package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const MessageUsersCode soul.ServerCode = 149

type MessageUsers struct{}

func (m MessageUsers) Serialize(usernames []string, message string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(MessageUsersCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(len(usernames)))
	if err != nil {
		return nil, err
	}

	for _, username := range usernames {
		err = internal.WriteString(buf, username)
		if err != nil {
			return nil, err
		}
	}

	err = internal.WriteString(buf, message)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
