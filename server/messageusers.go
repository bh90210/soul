package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const MessageUsersCode Code = 149

type MessageUsers struct{}

func (m MessageUsers) Serialize(usernames []string, message string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(MessageUsersCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(len(usernames)))
	if err != nil {
		return nil, err
	}

	for _, username := range usernames {
		err = soul.WriteString(buf, username)
		if err != nil {
			return nil, err
		}
	}

	err = soul.WriteString(buf, message)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
