package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeUploadFailed Code = 46

// UploadFailed code 46 this message is sent whenever a file connection of an
// active upload closes. Soulseek NS clients can also send this message when
// a file cannot be read. The recipient either re-queues the upload (download on
// their end), or ignores the message if the transfer finished.
type UploadFailed struct {
	Filename string
}

// Serialize accepts a UploadFailed and returns a message packed as a byte slice.
func (u *UploadFailed) Serialize(message *UploadFailed) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := internal.WriteUint32(buf, uint32(CodeUploadFailed))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Filename)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

// Deserialize populates a UploadFailed with the data in the provided reader.
func (u *UploadFailed) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 46
	if err != nil {
		return err
	}

	if code != uint32(CodeUploadFailed) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeUploadFailed, code))
	}

	u.Filename, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
