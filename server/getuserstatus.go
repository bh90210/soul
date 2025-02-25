package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const GetUserStatusCode soul.CodeServer = 7

type GetUserStatus struct {
	Username   string
	Status     UserStatus
	Privileged bool
}

// Serialize serializes the GetUserStatus struct into a byte slice
func (g GetUserStatus) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(GetUserStatusCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (g *GetUserStatus) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 7
	if err != nil {
		return err
	}

	if code != uint32(GetUserStatusCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", GetUserStatusCode, code))
	}

	g.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	status, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	g.Status = UserStatus(status)

	g.Privileged, err = internal.ReadBool(reader)
	if err != nil {
		return err
	}

	return nil
}
