package file

import (
	"bytes"
	"io"

	"github.com/bh90210/soul"
)

type Offset struct {
	Offset int
}

func (o Offset) Serialize(offset int) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := soul.WriteInt64(buf, int64(offset))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (o *Offset) Deserialize(reader io.Reader) (err error) {
	o.Offset, err = soul.ReadInt64ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
