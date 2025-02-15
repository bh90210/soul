package parentminspeed

import (
	"io"

	"github.com/bh90210/soul"
)

const Code soul.UInt = 83

func Deserialize(reader io.Reader) int {
	soul.ReadUInt(reader) // size
	soul.ReadUInt(reader) // code 83

	minSpeed := soul.ReadUInt(reader)

	return int(minSpeed)
}
