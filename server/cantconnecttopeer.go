package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeCantConnectToPeer Code = 1001

type CantConnectToPeer struct {
	Token    soul.Token
	Username string
}

func (c *CantConnectToPeer) Serialize(message *CantConnectToPeer) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeCantConnectToPeer))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Username)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (c *CantConnectToPeer) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 1001
	if err != nil {
		return err
	}

	if code != uint32(CodeCantConnectToPeer) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeCantConnectToPeer, code))
	}

	c.Token, err = internal.ReadUint32ToToken(reader)
	if err != nil {
		return err
	}

	c.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
