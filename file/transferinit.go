package file

import (
	"bytes"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

type TransferInit struct {
	Token soul.Token
}

func (t TransferInit) Serialize(token soul.Token) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := internal.WriteUint32(buf, token.Uint32())
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *TransferInit) Deserialize(reader io.Reader) (err error) {
	t.Token, err = internal.ReadUint32ToToken(reader)
	if err != nil {
		return
	}

	return
}
