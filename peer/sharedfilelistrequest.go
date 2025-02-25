package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const SharedFileListRequestCode soul.CodePeer = 4

type SharedFileListRequest struct{}

func (g SharedFileListRequest) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(SharedFileListRequestCode))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (g *SharedFileListRequest) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 4
	if err != nil {
		return err
	}

	if code != uint32(SharedFileListRequestCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", SharedFileListRequestCode, code))
	}

	return nil
}
