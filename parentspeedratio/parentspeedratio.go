package parentspeedratio

import (
	"io"

	"github.com/bh90210/soul"
)

const Code soul.UInt = 84

func Deserialize(reader io.Reader) int {
	soul.ReadUInt(reader) // size
	soul.ReadUInt(reader) // code 84

	speedRatio := soul.ReadUInt(reader)

	return int(speedRatio)
}
