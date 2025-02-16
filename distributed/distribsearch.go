package distributed

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bh90210/soul"
)

const DistribSearchCode Code = 3

type DistribSearch struct {
	Username string
	Token    int
	Query    string
}

func (d DistribSearch) Serialize(token int, username, query string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint8(buf, uint8(DistribSearchCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(0))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(token))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, query)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (d *DistribSearch) Deserialize(reader *bytes.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint8(reader) // code 3
	if err != nil {
		return err
	}

	if code != uint8(DistribSearchCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", DistribSearchCode, code))
	}

	_, err = soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	d.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	d.Token, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	d.Query, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
