package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeFileSearch Code = 26

type FileSearch struct {
	Username    string
	Token       soul.Token
	SearchQuery string
}

func (f *FileSearch) Serialize(message *FileSearch) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeFileSearch))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.SearchQuery)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (f *FileSearch) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 26
	if err != nil {
		return err
	}

	if code != uint32(CodeFileSearch) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeFileSearch, code))
	}

	f.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	f.Token, err = internal.ReadUint32ToToken(reader)
	if err != nil {
		return err
	}

	f.SearchQuery, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
