package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeUserInfoResponse Code = 16

type UserInfoResponse struct {
	Description     string
	Picture         []byte
	TotalUpload     uint32
	QueueSize       uint32
	FreeSlots       bool
	UploadPermitted UploadPermission
}

func (u *UserInfoResponse) Serialize(message *UserInfoResponse) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeUserInfoResponse))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Description)
	if err != nil {
		return nil, err
	}

	if message.Picture != nil {
		err = internal.WriteBool(buf, true)
		if err != nil {
			return nil, err
		}

		err = internal.WriteBytes(buf, message.Picture)
		if err != nil {
			return nil, err
		}
	} else {
		err = internal.WriteBool(buf, false)
		if err != nil {
			return nil, err
		}
	}

	err = internal.WriteUint32(buf, message.TotalUpload)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, message.QueueSize)
	if err != nil {
		return nil, err
	}

	err = internal.WriteBool(buf, message.FreeSlots)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.UploadPermitted))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (u *UserInfoResponse) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 16
	if err != nil {
		return err
	}

	if code != uint32(CodeUserInfoResponse) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeUserInfoResponse, code))
	}

	u.Description, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	hasPicture, err := internal.ReadBool(reader)
	if err != nil {
		return err
	}

	if hasPicture {
		u.Picture, err = internal.ReadBytes(reader)
		if err != nil {
			return err
		}
	}

	u.TotalUpload, err = internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	u.QueueSize, err = internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	u.FreeSlots, err = internal.ReadBool(reader)
	if err != nil {
		return err
	}

	var upload uint32
	upload, err = internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	u.UploadPermitted = UploadPermission(upload)

	return nil
}
