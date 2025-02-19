package server

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bh90210/soul"
)

const CantConnectToPeerCode Code = 1001

type CantConnectToPeer struct {
	Token    uint32
	Username string
}

func (c CantConnectToPeer) Serialize(token uint32, username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(CantConnectToPeerCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, token)
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
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 1001
	if err != nil {
		return err
	}

	if code != uint32(CantConnectToPeerCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CantConnectToPeerCode, code))
	}

	c.Token, err = soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	c.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
