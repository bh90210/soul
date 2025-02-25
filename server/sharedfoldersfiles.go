package server

import (
	"bytes"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const SharedFoldersFilesCode soul.CodeServer = 35

type SharedFoldersFiles struct{}

// Serialize accepts the number of directories and files and returns a serialized byte array.
func (SharedFoldersFiles) Serialize(directories, files int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(SharedFoldersFilesCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(directories))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(files))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
