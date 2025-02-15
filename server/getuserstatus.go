package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const GetUserStatusCode soul.UInt = 7

type GetUserStatus struct {
	Username   string
	Status     soul.UserStatusCode
	Privileged bool
}

// Serialize serializes the GetUserStatus struct into a byte slice
func (g GetUserStatus) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, GetUserStatusCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (g *GetUserStatus) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 7
	if err != nil {
		return err
	}

	if code != GetUserStatusCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", GetUserStatusCode, code))
	}

	g.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	status, err := soul.ReadUInt(reader)
	if err != nil {
		return err
	}

	g.Status = soul.UserStatusCode(status)

	g.Privileged, err = soul.ReadBool(reader)
	if err != nil {
		return err
	}

	return nil
}
