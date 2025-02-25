package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const SetStatusCode soul.CodeServer = 28

type SetStatus struct{}

func (s SetStatus) Serialize(status UserStatus) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(SetStatusCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(status))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
