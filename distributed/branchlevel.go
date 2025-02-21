package distributed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const BranchLevelCode Code = 4

type BranchLevel struct {
	Level int
}

func (d BranchLevel) Serialize(branchLevel int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint8(buf, uint8(BranchLevelCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(0))
	if err != nil {
		return nil, err
	}

	err = soul.WriteInt32(buf, int32(branchLevel))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (d *BranchLevel) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint8(reader) // code 4
	if err != nil {
		return err
	}

	if code != uint8(BranchLevelCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", BranchLevelCode, code))
	}

	d.Level, err = soul.ReadInt32ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
