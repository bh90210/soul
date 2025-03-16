package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodeMessageUsers Code = 149

type MessageUsers struct {
	Usernames []string
	Message   string
}

func (m *MessageUsers) Serialize(message *MessageUsers) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeMessageUsers))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(len(message.Usernames)))
	if err != nil {
		return nil, err
	}

	for _, username := range message.Usernames {
		err = internal.WriteString(buf, username)
		if err != nil {
			return nil, err
		}
	}

	err = internal.WriteString(buf, message.Message)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
