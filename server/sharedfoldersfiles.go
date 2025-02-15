package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const SharedFoldersFilesCode soul.UInt = 35

type SharedFoldersFiles struct{}

func (s SharedFoldersFiles) Serialize(directories, files int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, SharedFoldersFilesCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(directories))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(files))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
