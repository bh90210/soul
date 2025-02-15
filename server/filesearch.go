package server

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bh90210/soul"
)

const FileSearchCode soul.UInt = 26

type FileSearch struct {
	Username    string
	Token       int
	SearchQuery string
}

func (f FileSearch) Serialize(token int, searchQuery string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, FileSearchCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(token))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, searchQuery)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (f *FileSearch) Deserialize(reader *bytes.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 26
	if err != nil {
		return err
	}

	if code != FileSearchCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", FileSearchCode, code))
	}

	f.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	f.Token, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	f.SearchQuery, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
