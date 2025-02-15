package privilegedusers

import (
	"io"

	"github.com/bh90210/soul"
)

const Code soul.UInt = 69

func Deserialize(reader io.Reader) []string {
	soul.ReadUInt(reader) // size
	soul.ReadUInt(reader) // code 69

	numberOfUsers := soul.ReadUInt(reader)

	users := make([]string, 0)
	for i := 0; i < int(numberOfUsers); i++ {
		user := soul.ReadString(reader)

		users = append(users, user)
	}

	return users
}
