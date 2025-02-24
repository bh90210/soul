package distributed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const BranchLevelCode soul.DistributedCode = 4

type BranchLevel struct {
	Level int32
}

func (d BranchLevel) Serialize(branchLevel int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint8(buf, uint8(BranchLevelCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteInt32(buf, branchLevel)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (d *BranchLevel) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint8(reader) // code 4
	if err != nil {
		return err
	}

	if code != uint8(BranchLevelCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", BranchLevelCode, code))
	}

	d.Level, err = internal.ReadInt32(reader)
	if err != nil {
		return err
	}

	return nil
}
