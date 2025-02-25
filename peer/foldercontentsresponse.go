package peer

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// FolderContentsResponse code 37 peer responds with the contents of a
// particular folder (with all subfolders) after we’ve sent a FolderContentsRequest.
type FolderContentsResponse struct {
	Token   soul.Token
	Folder  string
	Folders []Folder
}

// Folder represents a folder and its contents.
type Folder struct {
	Directory string
	Files     []File
}

func (FolderContentsResponse) Serialize(token soul.Token, folder string, folders []Folder) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeFolderContentsResponse))
	if err != nil {
		return nil, err
	}

	gzw := gzip.NewWriter(buf)

	err = internal.WriteUint32(gzw, uint32(token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(gzw, folder)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(gzw, uint32(len(folders)))
	if err != nil {
		return nil, err
	}

	for _, f := range folders {
		err = internal.WriteString(gzw, f.Directory)
		if err != nil {
			return nil, err
		}

		err = internal.WriteUint32(gzw, uint32(len(f.Files)))
		if err != nil {
			return nil, err
		}

		for _, file := range f.Files {
			err = internal.WriteUint8(gzw, uint8(1))
			if err != nil {
				return nil, err
			}

			err = internal.WriteString(gzw, file.Name)
			if err != nil {
				return nil, err
			}

			err = internal.WriteUint64(gzw, file.Size)
			if err != nil {
				return nil, err
			}

			err = internal.WriteString(gzw, file.Extension)
			if err != nil {
				return nil, err
			}

			err = internal.WriteUint32(gzw, uint32(len(file.Attributes)))
			if err != nil {
				return nil, err
			}

			for _, attribute := range file.Attributes {
				err = internal.WriteUint32(gzw, uint32(attribute.Code))
				if err != nil {
					return nil, err
				}

				err = internal.WriteUint32(gzw, uint32(attribute.Value))
				if err != nil {
					return nil, err
				}
			}
		}
	}

	err = gzw.Close()
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

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

	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}

	defer gzr.Close()

	f.Token, err = internal.ReadUint32ToToken(gzr)
	if err != nil {
		return err
	}

	f.Folder, err = internal.ReadString(gzr)
	if err != nil {
		return err
	}

	folders, err := internal.ReadUint32(gzr)
	if err != nil {
		return err
	}

	for i := 0; i < int(folders); i++ {
		var folder Folder

		folder.Directory, err = internal.ReadString(gzr)
		if err != nil {
			return err
		}

		files, err := internal.ReadUint32(gzr)
		if err != nil {
			return err
		}

		for j := 0; j < int(files); j++ {
			var file File

			_, err = internal.ReadUint8(gzr)
			if err != nil {
				return err
			}

			file.Name, err = internal.ReadString(gzr)
			if err != nil {
				return err
			}

			file.Size, err = internal.ReadUint64(gzr)
			if err != nil {
				return err
			}

			file.Extension, err = internal.ReadString(gzr)
			if err != nil {
				return err
			}

			attributes, err := internal.ReadUint32(gzr)
			if err != nil {
				return err
			}

			for k := 0; k < int(attributes); k++ {
				var a Attribute

				code, err := internal.ReadUint32(gzr)
				if err != nil {
					return err
				}

				a.Code = FileAttributeType(code)

				a.Value, err = internal.ReadUint32(gzr)
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
