package file

import (
	"bytes"

	"github.com/bh90210/soul"
)

type TransferInit struct {
	Token int
}

func (t TransferInit) Serialize(token int) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := soul.WriteUint32(buf, uint32(token))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *TransferInit) Deserialize(reader *bytes.Reader) (err error) {
	t.Token, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return
	}

	return
}
