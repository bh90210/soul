package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const MessageAckedCode soul.ServerCode = 23

type MessageAcked struct{}

func (m MessageAcked) Serialize(messageID int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(MessageAckedCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(messageID))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
