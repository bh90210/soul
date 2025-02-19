package file

import (
	"bytes"

	"github.com/bh90210/soul"
)

type TransferInit struct {
	Token uint32
}

func (t TransferInit) Serialize(token uint32) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := soul.WriteUint32(buf, token)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *TransferInit) Deserialize(reader *bytes.Reader) (err error) {
	t.Token, err = soul.ReadUint32(reader)
	if err != nil {
		return
	}

	return
}
