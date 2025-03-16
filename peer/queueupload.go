package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeQueueUpload Code = 43

// QueueUpload code 43 this message is used to tell a peer that an upload should be queued
// on their end. Once the recipient is ready to transfer the requested file,
// they will send a TransferRequest to us.
type QueueUpload struct {
	Filename string
}

func (q *QueueUpload) Serialize(message *QueueUpload) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := internal.WriteUint32(buf, uint32(CodeQueueUpload))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Filename)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (q *QueueUpload) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 43
	if err != nil {
		return err
	}

	if code != uint32(CodeQueueUpload) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeQueueUpload, code))
	}

	q.Filename, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
