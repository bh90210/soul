package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const MessageAckedCode soul.UInt = 23

type MessageAcked struct{}

func (m MessageAcked) Serialize(messageID int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, MessageAckedCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(messageID))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
