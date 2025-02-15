package server

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bh90210/soul"
)

const CantConnectToPeerCode soul.UInt = 1001

type CantConnectToPeer struct {
	Token    int
	Username string
}

func (c CantConnectToPeer) Serialize(token int, username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, CantConnectToPeerCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteUInt(buf, soul.UInt(token))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (c *CantConnectToPeer) Deserialize(reader *bytes.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 1001
	if err != nil {
		return err
	}

	if code != CantConnectToPeerCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CantConnectToPeerCode, code))
	}

	c.Token, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	c.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
