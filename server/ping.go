package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const PingCode Code = 32

type Ping struct{}

func (p Ping) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(PingCode))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
