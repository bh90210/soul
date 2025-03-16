package server

import (
	"bytes"

	"github.com/bh90210/soul/internal"
)

const CodeSharedFoldersFiles Code = 35

type SharedFoldersFiles struct {
	Directories int
	Files       int
}

// Serialize accepts the number of directories and files and returns a serialized byte array.
func (s *SharedFoldersFiles) Serialize(message *SharedFoldersFiles) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeSharedFoldersFiles))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Directories))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Files))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}
