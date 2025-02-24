package distributed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const EmbeddedMessageCode soul.DistributedCode = 93

type EmbeddedMessage struct {
	Code    soul.DistributedCode
	Message []byte
}

func (d EmbeddedMessage) Serialize(code soul.DistributedCode, message []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint8(buf, uint8(EmbeddedMessageCode))
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

func (d *EmbeddedMessage) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint8(reader) // code 93
	if err != nil {
		return err
	}

	if code != uint8(EmbeddedMessageCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", EmbeddedMessageCode, code))
	}

	code, err = internal.ReadUint8(reader)
	if err != nil {
		return err
	}

	d.Code = soul.DistributedCode(code)

	d.Message, err = internal.ReadBytes(reader)
	if err != nil {
		return err
	}

	return nil
}
