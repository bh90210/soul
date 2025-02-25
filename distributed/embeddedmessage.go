package distributed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// EmbeddedMessage code 93 a branch root sends us an embedded distributed message. We unpack the
// distributed message and distribute it to our child peers. The only type of distributed
// message sent at present is DistribSearch (distributed code 3).
type EmbeddedMessage struct {
	Code    soul.CodeDistributed
	Message []byte
}

// Serialize accepts a code and message and returns a message packed as a byte slice.
func (EmbeddedMessage) Serialize(code soul.CodeDistributed, message []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint8(buf, uint8(CodeEmbeddedMessage))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint8(buf, uint8(code))
	if err != nil {
		return nil, err
	}

	err = internal.WriteBytes(buf, message)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

// Deserialize accepts a reader and deserializes the message into the EmbeddedMessage struct.
func (e *EmbeddedMessage) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint8(reader) // code 93
	if err != nil {
		return err
	}

	if code != uint8(CodeEmbeddedMessage) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeEmbeddedMessage, code))
	}

	code, err = internal.ReadUint8(reader)
	if err != nil {
		return err
	}

	e.Code = soul.CodeDistributed(code)

	e.Message, err = internal.ReadBytes(reader)
	if err != nil {
		return err
	}

	return nil
}
