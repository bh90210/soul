package distributed

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bh90210/soul"
)

const EmbeddedMessageCode Code = 93

type EmbeddedMessage struct {
	Code    Code
	Message []byte
}

func (d EmbeddedMessage) Serialize(code Code, message []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint8(buf, uint8(EmbeddedMessageCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(code))
	if err != nil {
		return nil, err
	}

	err = soul.WriteBytes(buf, message)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (d *EmbeddedMessage) Deserialize(reader *bytes.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint8(reader) // code 93
	if err != nil {
		return err
	}

	if code != uint8(EmbeddedMessageCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", EmbeddedMessageCode, code))
	}

	code, err = soul.ReadUint8(reader)
	if err != nil {
		return err
	}

	d.Code = Code(code)

	d.Message, err = soul.ReadBytes(reader)
	if err != nil {
		return err
	}

	return nil
}
