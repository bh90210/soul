package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodePrivateRoomDisown Code = 137

type PrivateRoomDisown struct{}

func (p PrivateRoomDisown) Serialize(room string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodePrivateRoomDisown))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
