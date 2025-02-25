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

// SharedFileListResponse code 5 peer responds with a list of shared
// files after we’ve sent a SharedFileListRequest.
type SharedFileListResponse struct {
	Directories        []Directory
	PrivateDirectories []Directory
}

// Directory is a directory in a shared file list.
type Directory struct {
	Name  string
	Files []File
}

// File is a file in a directory.
type File struct {
	Name       string
	Size       uint64
	Extension  string
	Attributes []Attribute
}

// Attribute is a type of file attribute.
type Attribute struct {
	Code  FileAttributeType
	Value uint32
}

// Serialize accepts directories and privateDirectories and returns a message packed as a byte slice.
// It uses custom errors for the following cases:
// - ErrNoDirectories is returned when there are no directories.
// - ErrEmptyDirectoryName is returned when the directory name is empty.
// - ErrEmptyDirectory is returned when the directory is empty.
// - ErrEmptyFileName is returned when the file name is empty.
// - ErrSizeZero is returned when the file size is zero.
// - ErrEmptyFileExtension is returned when the file extension is empty.
// You can use them in your code to check for specific errors.
func (s SharedFileListResponse) Serialize(directories, privateDirectories []Directory) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeSharedFileListResponse))
	if err != nil {
		return nil, err
	}

	gzw := gzip.NewWriter(buf)

	err = s.walkWrite(directories, gzw)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(gzw, 0)
	if err != nil {
		return nil, err
	}

	err = s.walkWrite(privateDirectories, gzw)
	if err != nil {
		return nil, err
	}

	err = gzw.Close()
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

// ErrNoDirectories is returned when there are no directories.
var ErrNoDirectories = errors.New("no directories")

// ErrEmptyDirectoryName is returned when the directory name is empty.
var ErrEmptyDirectoryName = errors.New("directory name is empty")

// ErrEmptyDirectory is returned when the directory is empty.
var ErrEmptyDirectory = errors.New("directory is empty")

// ErrEmptyFileName is returned when the file name is empty.
var ErrEmptyFileName = errors.New("file name is empty")

// ErrSizeZero is returned when the file size is zero.
var ErrSizeZero = errors.New("file size is zero")

// ErrEmptyFileExtension is returned when the file extension is empty.
var ErrEmptyFileExtension = errors.New("file extension is empty")

func (SharedFileListResponse) walkWrite(directories []Directory, gzw *gzip.Writer) error {
	err := internal.WriteUint32(gzw, uint32(len(directories)))
	if err != nil {
		return err
	}

	for _, directory := range directories {
		if directory.Name == "" {
			return ErrEmptyDirectoryName
		}

		err = internal.WriteString(gzw, directory.Name)
		if err != nil {
			return err
		}

		if len(directory.Files) == 0 {
			return ErrEmptyDirectory
		}

		err = internal.WriteUint32(gzw, uint32(len(directory.Files)))
		if err != nil {
			return err
		}

		for _, file := range directory.Files {
			if file.Name == "" {
				return ErrEmptyFileName
			}

			err = internal.WriteUint8(gzw, 1)
			if err != nil {
				return err
			}

			err = internal.WriteString(gzw, file.Name)
			if err != nil {
				return err
			}

			if file.Size == 0 {
				return ErrSizeZero
			}

			err = internal.WriteUint64(gzw, file.Size)
			if err != nil {
				return err
			}

			if file.Extension == "" {
				return ErrEmptyFileExtension
			}

			err = internal.WriteString(gzw, file.Extension)
			if err != nil {
				return err
			}

			err = internal.WriteUint32(gzw, uint32(len(file.Attributes)))
			if err != nil {
				return err
			}

			for _, attribute := range file.Attributes {
				err = internal.WriteUint32(gzw, uint32(attribute.Code))
				if err != nil {
					return err
				}

				err = internal.WriteUint32(gzw, attribute.Value)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Deserialize accepts a reader and deserializes the message into the SharedFileListResponse struct.
func (s *SharedFileListResponse) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 5
	if err != nil {
		return err
	}

	if code != uint32(CodeSharedFileListResponse) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeSharedFileListResponse, code))
	}

	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}

	defer gzr.Close()

	directories, err := internal.ReadUint32(gzr)
	if err != nil {
		return err
	}

	s.Directories, err = s.walkRead(directories, gzr)
	if err != nil {
		return err
	}

	_, err = internal.ReadUint32(gzr)
	if err != nil {
		return err
	}

	privateDirectories, err := internal.ReadUint32(gzr)
	if err != nil {
		return err
	}

	s.PrivateDirectories, err = s.walkRead(privateDirectories, gzr)
	if err != nil {
		return err
	}

	return nil
}

func (s *SharedFileListResponse) walkRead(numberOfDirectories uint32, gzr *gzip.Reader) (directories []Directory, err error) {
	for i := 0; i < int(numberOfDirectories); i++ {
		var directory Directory
		var err error

		directory.Name, err = internal.ReadString(gzr)
		if err != nil {
			return nil, err
		}

		files, err := internal.ReadUint32(gzr)
		if err != nil {
			return nil, err
		}

		for j := 0; j < int(files); j++ {
			var f File

			_, err := internal.ReadUint8(gzr)
			if err != nil {
				return nil, err
			}

			f.Name, err = internal.ReadString(gzr)
			if err != nil {
				return nil, err
			}

			f.Size, err = internal.ReadUint64(gzr)
			if err != nil {
				return nil, err
			}

			f.Extension, err = internal.ReadString(gzr)
			if err != nil {
				return nil, err
			}

			attributes, err := internal.ReadUint32(gzr)
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			}

			for k := 0; k < int(attributes); k++ {
				var a Attribute

				code, err := internal.ReadUint32(gzr)
				if err != nil {
					return nil, err
				}

				a.Code = FileAttributeType(code)

				a.Value, err = internal.ReadUint32(gzr)
				if err != nil && !errors.Is(err, io.EOF) {
					return nil, err
				}

				f.Attributes = append(f.Attributes, a)
			}

			directory.Files = append(directory.Files, f)
		}

		directories = append(directories, directory)
	}

	return directories, err
}
