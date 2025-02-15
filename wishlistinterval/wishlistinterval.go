package wishlistinterval

import (
	"io"

	"github.com/bh90210/soul"
)

const Code soul.UInt = 104

func Deserialize(reader io.Reader) int {
	soul.ReadUInt(reader) // size
	soul.ReadUInt(reader) // code 104

	interval := soul.ReadUInt(reader)

	return int(interval)
}
