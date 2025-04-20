package peer

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeFolderContentsResponse Code = 37

// FolderContentsResponse code 37 peer responds with the contents of a
// particular folder (with all subfolders) after we’ve sent a FolderContentsRequest.
type FolderContentsResponse struct {
	Token   soul.Token
	Folder  string
	Folders []Directory
}

// Serialize accepts a FolderContentsResponse and returns a message packed as a byte slice.
func (f *FolderContentsResponse) Serialize(message *FolderContentsResponse) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeFolderContentsResponse))
	if err != nil {
		return nil, err
	}

	zw := zlib.NewWriter(buf)

	err = internal.WriteUint32(zw, uint32(message.Token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(zw, message.Folder)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(zw, uint32(len(message.Folders)))
	if err != nil {
		return nil, err
	}

	for _, f := range message.Folders {
		err = internal.WriteString(zw, f.Name)
		if err != nil {
			return nil, err
		}

		err = internal.WriteUint32(zw, uint32(len(f.Files)))
		if err != nil {
			return nil, err
		}

		for _, file := range f.Files {
			err = internal.WriteUint8(zw, uint8(1))
			if err != nil {
				return nil, err
			}

			err = internal.WriteString(zw, file.Name)
			if err != nil {
				return nil, err
			}

			err = internal.WriteUint64(zw, file.Size)
			if err != nil {
				return nil, err
			}

			err = internal.WriteString(zw, file.Extension)
			if err != nil {
				return nil, err
			}

			err = internal.WriteUint32(zw, uint32(len(file.Attributes)))
			if err != nil {
				return nil, err
			}

			for _, attribute := range file.Attributes {
				err = internal.WriteUint32(zw, uint32(attribute.Code))
				if err != nil {
					return nil, err
				}

				err = internal.WriteUint32(zw, uint32(attribute.Value))
				if err != nil {
					return nil, err
				}
			}
		}
	}

	err = zw.Close()
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

// Deserialize populates a FolderContentsResponse with the data in the provided reader.
func (f *FolderContentsResponse) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 37
	if err != nil {
		return err
	}

	if code != uint32(CodeFolderContentsResponse) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeFolderContentsResponse, code))
	}

	zr, err := zlib.NewReader(reader)
	if err != nil {
		return err
	}

	defer zr.Close()

	f.Token, err = internal.ReadUint32ToToken(zr)
	if err != nil {
		return err
	}

	f.Folder, err = internal.ReadString(zr)
	if err != nil {
		return err
	}

	folders, err := internal.ReadUint32(zr)
	if err != nil {
		return err
	}

	for range int(folders) {
		var folder Directory

		folder.Name, err = internal.ReadString(zr)
		if err != nil {
			return err
		}

		files, err := internal.ReadUint32(zr)
		if err != nil {
			return err
		}

		for range int(files) {
			var file File

			_, err = internal.ReadUint8(zr)
			if err != nil {
				return err
			}

			file.Name, err = internal.ReadString(zr)
			if err != nil {
				return err
			}

			file.Size, err = internal.ReadUint64(zr)
			if err != nil {
				return err
			}

			file.Extension, err = internal.ReadString(zr)
			if err != nil {
				return err
			}

			attributes, err := internal.ReadUint32(zr)
			if err != nil {
				return err
			}

			for range int(attributes) {
				var a Attribute

				code, err := internal.ReadUint32(zr)
				if err != nil {
					return err
				}

				a.Code = FileAttributeType(code)

				a.Value, err = internal.ReadUint32(zr)
				if err != nil {
					return err
				}

				file.Attributes = append(file.Attributes, a)
			}

			folder.Files = append(folder.Files, file)
		}

		f.Folders = append(f.Folders, folder)
	}

	return nil
}
