package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodeSharedFoldersFiles Code = 35

type SharedFoldersFiles struct{}

// Serialize accepts the number of directories and files and returns a serialized byte array.
func (SharedFoldersFiles) Serialize(directories, files int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeSharedFoldersFiles))
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
