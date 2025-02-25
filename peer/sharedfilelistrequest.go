package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

type SharedFileListRequest struct{}

func (SharedFileListRequest) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeSharedFileListRequest))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (s *SharedFileListRequest) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 4
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	if code != uint32(CodeSharedFileListRequest) {
		return errors.Join(err, soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeSharedFileListRequest, code))
	}

	return err
}
