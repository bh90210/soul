package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeUserInfoRequest Code = 15

// UserInfoRequest code 15, we send this to ask the other peer to send us their user information, picture and all.
type UserInfoRequest struct{}

// Serialize accepts a UserInfoRequest and returns a message packed as a byte slice.
func (u *UserInfoRequest) Serialize(_ *UserInfoRequest) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeUserInfoRequest))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

// Deserialize populates a UserInfoRequest with the data in the provided reader.
func (u *UserInfoRequest) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 5
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	if code != uint32(CodeUserInfoRequest) {
		return errors.Join(err, soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeUserInfoRequest, code))
	}

	return err
}
