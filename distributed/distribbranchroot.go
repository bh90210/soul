package distributed

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bh90210/soul"
)

const DistribBranchRootCode Code = 5

type DistribBranchRoot struct {
	Root string
}

func (d DistribBranchRoot) Serialize(root string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint8(buf, uint8(DistribBranchRootCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, root)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (d *DistribBranchRoot) Deserialize(reader *bytes.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint8(reader) // code 5
	if err != nil {
		return err
	}

	if code != uint8(DistribBranchRootCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", DistribBranchRootCode, code))
	}

	d.Root, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
