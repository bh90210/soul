package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const GivePrivilegesCode soul.CodeServer = 123

type GivePrivileges struct{}

func (g GivePrivileges) Serialize(username string, days int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(GivePrivilegesCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(days))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
