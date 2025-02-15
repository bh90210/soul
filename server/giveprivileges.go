package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const GivePrivilegesCode soul.UInt = 123

type GivePrivileges struct{}

func (g GivePrivileges) Serialize(username string, days int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, GivePrivilegesCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(days))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
