package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// UploadDenied code 50 this message is sent to reject QueueUpload attempts
// and previously queued files. The reason for rejection will appear in the
// transfer list of the recipient.
type UploadDenied struct {
	Filename string
	Reason   error
}

func (UploadDenied) Serialize(filename string, reason error) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := internal.WriteUint32(buf, uint32(CodeUploadDenied))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, filename)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, reason.Error())
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (u *UploadDenied) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 50
	if err != nil {
		return err
	}

	if code != uint32(CodeUploadDenied) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeUploadDenied, code))
	}

	u.Filename, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	r, err := internal.ReadString(reader)
	if err != nil {
		return err
	}

	u.Reason = reason(r)

	return nil
}
