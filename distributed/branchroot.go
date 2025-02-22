package distributed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const BranchRootCode soul.DistributedCode = 5

type BranchRoot struct {
	Root string
}

func (d BranchRoot) Serialize(root string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint8(buf, uint8(BranchRootCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, root)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (d *BranchRoot) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint8(reader) // code 5
	if err != nil {
		return err
	}

	if code != uint8(BranchRootCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", BranchRootCode, code))
	}

	d.Root, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
