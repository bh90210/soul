package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const MessageAckedCode soul.ServerCode = 23

type MessageAcked struct{}

func (m MessageAcked) Serialize(messageID int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(MessageAckedCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(messageID))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
