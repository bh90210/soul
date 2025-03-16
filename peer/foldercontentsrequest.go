package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeFolderContentsRequest Code = 36

// FolderContentsRequest code 36 we ask the peer to send us the contents of a single folder.
type FolderContentsRequest struct {
	Token  soul.Token
	Folder string
}

func (f *FolderContentsRequest) Serialize(message *FolderContentsRequest) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeFolderContentsRequest))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Folder)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (f *FolderContentsRequest) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 36
	if err != nil {
		return err
	}

	if code != uint32(CodeFolderContentsRequest) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeFolderContentsRequest, code))
	}

	f.Token, err = internal.ReadUint32ToToken(reader)
	if err != nil {
		return err
	}

	f.Folder, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
