package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomCancelMembership Code = 136

type PrivateRoomCancelMembership struct {
	Room string
}

func (p *PrivateRoomCancelMembership) Serialize(message *PrivateRoomCancelMembership) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodePrivateRoomCancelMembership))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Room)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
