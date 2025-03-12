package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeTransferResponse Code = 41

// TransferResponse code 41 response to TransferRequest.
// We (or the other peer) either agrees, or tells the reason
// for rejecting the file upload.
type TransferResponse struct {
	Token   soul.Token
	Allowed bool
	Reason  error
}

// ErrNotAllowedWithNoReason is returned when a TransferResponse is not allowed and no reason is provided.
var ErrNotAllowedWithNoReason = errors.New("rejection reason is required when transfer is not allowed")

// Serialize accepts a TransferResponse and returns a message packed as a byte slice.
// If the transfer is not allowed, a reason must be provided. The possible errors are:
// ErrBanned, ErrCancelled, ErrComplete, ErrFileNotShared, ErrFileReadError, ErrPendingShutdown,
// ErrQueued, ErrTooManyFiles, ErrTooManyMegabytes, and ErrNotAllowedWithNoReason.
// All errors exist in the peer.TransferResponse package.
func (TransferResponse) Serialize(token soul.Token, allowed bool, reason ...error) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := internal.WriteUint32(buf, uint32(CodeTransferResponse))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteBool(buf, allowed)
	if err != nil {
		return nil, err
	}

	if !allowed {
		if len(reason) == 0 {
			return nil, ErrNotAllowedWithNoReason
		}

		err = internal.WriteString(buf, reason[0].Error())
		if err != nil {
			return nil, err
		}
	}

	return internal.Pack(buf.Bytes())
}

func (t *TransferResponse) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 41
	if err != nil {
		return err
	}

	if code != uint32(CodeTransferResponse) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeTransferResponse, code))
	}

	token, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	t.Token = soul.Token(token)

	t.Allowed, err = internal.ReadBool(reader)
	if err != nil {
		return err
	}

	if !t.Allowed {
		r, err := internal.ReadString(reader)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		t.Reason = Reason(r)
	}

	return err
}
