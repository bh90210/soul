package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const FileSearchCode soul.ServerCode = 26

type FileSearch struct {
	Username    string
	Token       soul.Token
	SearchQuery string
}

func (f FileSearch) Serialize(token soul.Token, searchQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(FileSearchCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, searchQuery)
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

	if code != uint32(FileSearchCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", FileSearchCode, code))
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
