package distributed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// BranchLevel code 4 we tell our distributed children what our position is in our branch (xth generation)
// on the distributed network. If we receive a branch level of 0 from a parent, we should
// mark the parent as our branch root, since they won’t send a DistribBranchRoot message
// in this case.
type BranchLevel struct {
	Level int32
}

// Serialize accepts a branch level and returns a message packed as a byte slice.
func (d BranchLevel) Serialize(branchLevel int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint8(buf, uint8(CodeBranchLevel))
	if err != nil {
		return nil, err
	}

	err = internal.WriteInt32(buf, branchLevel)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

// Deserialize accepts a reader and deserializes the message into the BranchLevel struct.
func (d *BranchLevel) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint8(reader) // code 4
	if err != nil {
		return err
	}

	if code != uint8(CodeBranchLevel) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeBranchLevel, code))
	}

	d.Level, err = internal.ReadInt32(reader)
	if err != nil {
		return err
	}

	return nil
}
