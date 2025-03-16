package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodeGivePrivileges Code = 123

type GivePrivileges struct {
	Username string
	Days     int
}

func (g *GivePrivileges) Serialize(message *GivePrivileges) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeGivePrivileges))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Days))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
