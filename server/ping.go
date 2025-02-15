package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const PingCode soul.UInt = 32

type Ping struct{}

func (p Ping) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, PingCode)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
