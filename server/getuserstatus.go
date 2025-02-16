package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const GetUserStatusCode Code = 7

type GetUserStatus struct {
	Username   string
	Status     soul.UserStatus
	Privileged bool
}

// Serialize serializes the GetUserStatus struct into a byte slice
func (g GetUserStatus) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(GetUserStatusCode))
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
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 7
	if err != nil {
		return err
	}

	if code != uint32(GetUserStatusCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", GetUserStatusCode, code))
	}

	g.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	status, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	g.Status = soul.UserStatus(status)

	g.Privileged, err = soul.ReadBool(reader)
	if err != nil {
		return err
	}

	return nil
}
