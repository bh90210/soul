package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodeUnwatchUser Code = 6

type UnwatchUser struct{}

func (u UnwatchUser) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeUnwatchUser))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
