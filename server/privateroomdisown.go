package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomDisown Code = 137

type PrivateRoomDisown struct {
	Room string
}

func (p *PrivateRoomDisown) Serialize(message *PrivateRoomDisown) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodePrivateRoomDisown))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Room)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
