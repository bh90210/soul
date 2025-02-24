package distributed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// BranchRootCode 5.
const BranchRootCode soul.DistributedCode = 5

// BranchRoot we tell our distributed children the username of the root of the
// branch we’re in on the distributed network. This message should not be sent
// when we’re the branch root.
type BranchRoot struct {
	Root string
}

// Serialize accepts a root and returns a message packed as a byte slice.
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

// Deserialize accepts a reader and deserializes the message into the BranchRoot struct.
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
