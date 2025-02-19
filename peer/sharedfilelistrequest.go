package peer

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bh90210/soul"
)

const SharedFileListRequestCode Code = 4

type SharedFileListRequest struct{}

func (g SharedFileListRequest) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(SharedFileListRequestCode))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (g *SharedFileListRequest) Deserialize(reader *bytes.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 4
	if err != nil {
		return err
	}

	if code != uint32(SharedFileListRequestCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", SharedFileListRequestCode, code))
	}

	return nil
}
