package distributed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const SearchCode soul.DistributedCode = 3

type Search struct {
	Username string
	Token    soul.Token
	Query    string
}

func (d Search) Serialize(token soul.Token, username, query string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint8(buf, uint8(SearchCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(0))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, token.Uint32())
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, query)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (d *Search) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint8(reader) // code 3
	if err != nil {
		return err
	}

	if code != uint8(SearchCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", SearchCode, code))
	}

	_, err = internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	d.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	d.Token, err = internal.ReadUint32ToToken(reader)
	if err != nil {
		return err
	}

	d.Query, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
