package port

import (
	"bytes"

	soul "github.com/bh90210/soul"
)

const Code soul.UInt = 2

func Write(port soul.UInt) []byte {
	buf := new(bytes.Buffer)
	soul.WriteUInt(buf, Code)
	return soul.Pack(buf.Bytes())
}
