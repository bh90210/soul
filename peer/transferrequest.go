package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeTransferRequest Code = 40

// TransferRequest code 40 this message is sent by a peer once they are ready to start
// uploading a file to us. A TransferResponse message is expected from the recipient,
// either allowing or rejecting the upload attempt.
//
// This message was formerly used to send a download request (direction 0) as well,
// but Nicotine+ >= 3.0.3, Museek+ and the official clients use the
// QueueUpload peer message for this purpose today.
type TransferRequest struct {
	Direction TransferDirection
	Token     soul.Token
	Filename  string
	FileSize  uint64
}

// Serialize accepts a TransferRequest and returns a message packed as a byte slice.
func (t *TransferRequest) Serialize(message *TransferRequest) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := internal.WriteUint32(buf, uint32(CodeTransferRequest))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Direction))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Filename)
	if err != nil {
		return nil, err
	}

	if message.Direction == UploadToPeer {
		err = internal.WriteUint64(buf, message.FileSize)
		if err != nil {
			return nil, err
		}
	}

	return internal.Pack(buf.Bytes())
}

// Deserialize populates a TransferRequest with the data in the provided reader.
func (t *TransferRequest) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 40
	if err != nil {
		return err
	}

	if code != uint32(CodeTransferRequest) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeTransferRequest, code))
	}

	direction, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	t.Direction = TransferDirection(direction)

	t.Token, err = internal.ReadUint32ToToken(reader)
	if err != nil {
		return err
	}

	t.Filename, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	if t.Direction == UploadToPeer {
		t.FileSize, err = internal.ReadUint64(reader)
		if err != nil {
			return err
		}
	}

	return nil
}
