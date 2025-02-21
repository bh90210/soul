package file

import (
	"bytes"
	"io"

	"github.com/bh90210/soul"
)

type Offset struct {
	Offset uint64
}

func (o Offset) Serialize(offset uint64) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint64(buf, offset)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (o *Offset) Deserialize(reader io.Reader) (err error) {
	o.Offset, err = soul.ReadUint64(reader)
	if err != nil {
		return err
	}

	return nil
}
