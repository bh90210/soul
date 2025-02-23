package file

import (
	"bytes"
	"io"

	"github.com/bh90210/soul/internal"
)

type TransferInit struct {
	Token uint32
}

func (t TransferInit) Serialize(token uint32) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := internal.WriteUint32(buf, token)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *TransferInit) Deserialize(reader io.Reader) (err error) {
	t.Token, err = internal.ReadUint32(reader)
	if err != nil {
		return
	}

	return
}
