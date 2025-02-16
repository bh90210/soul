package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const SharedFoldersFilesCode Code = 35

type SharedFoldersFiles struct{}

func (s SharedFoldersFiles) Serialize(directories, files int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(SharedFoldersFilesCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(directories))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(files))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
