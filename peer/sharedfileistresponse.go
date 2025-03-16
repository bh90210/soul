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

const CodeSharedFileListResponse Code = 5

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
func (s *SharedFileListResponse) Serialize(message *SharedFileListResponse) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeSharedFileListResponse))
	if err != nil {
		return nil, err
	}

	zw := zlib.NewWriter(buf)

	err = s.walkWrite(message.Directories, zw)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(zw, 0)
	if err != nil {
		return nil, err
	}

	err = s.walkWrite(message.PrivateDirectories, zw)
	if err != nil {
		return nil, err
	}

	err = zw.Close()
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

func (SharedFileListResponse) walkWrite(directories []Directory, zw *zlib.Writer) error {
	err := internal.WriteUint32(zw, uint32(len(directories)))
	if err != nil {
		return err
	}

	for _, directory := range directories {
		if directory.Name == "" {
			return ErrEmptyDirectoryName
		}

		err = internal.WriteString(zw, directory.Name)
		if err != nil {
			return err
		}

		if len(directory.Files) == 0 {
			return ErrEmptyDirectory
		}

		err = internal.WriteUint32(zw, uint32(len(directory.Files)))
		if err != nil {
			return err
		}

		for _, file := range directory.Files {
			if file.Name == "" {
				return ErrEmptyFileName
			}

			err = internal.WriteUint8(zw, 1)
			if err != nil {
				return err
			}

			err = internal.WriteString(zw, file.Name)
			if err != nil {
				return err
			}

			if file.Size == 0 {
				return ErrSizeZero
			}

			err = internal.WriteUint64(zw, file.Size)
			if err != nil {
				return err
			}

			if file.Extension == "" {
				return ErrEmptyFileExtension
			}

			err = internal.WriteString(zw, file.Extension)
			if err != nil {
				return err
			}

			err = internal.WriteUint32(zw, uint32(len(file.Attributes)))
			if err != nil {
				return err
			}

			for _, attribute := range file.Attributes {
				err = internal.WriteUint32(zw, uint32(attribute.Code))
				if err != nil {
					return err
				}

				err = internal.WriteUint32(zw, attribute.Value)
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

	zr, err := zlib.NewReader(reader)
	if err != nil {
		return err
	}

	defer zr.Close()

	directories, err := internal.ReadUint32(zr)
	if err != nil {
		return err
	}

	s.Directories, err = s.walkRead(directories, zr)
	if err != nil {
		return err
	}

	_, err = internal.ReadUint32(zr)
	if err != nil {
		return err
	}

	privateDirectories, err := internal.ReadUint32(zr)
	if err != nil {
		return err
	}

	s.PrivateDirectories, err = s.walkRead(privateDirectories, zr)
	if err != nil {
		return err
	}

	return nil
}

func (s *SharedFileListResponse) walkRead(numberOfDirectories uint32, zr io.ReadCloser) (directories []Directory, err error) {
	for range int(numberOfDirectories) {
		var directory Directory
		var err error

		directory.Name, err = internal.ReadString(zr)
		if err != nil {
			return nil, err
		}

		files, err := internal.ReadUint32(zr)
		if err != nil {
			return nil, err
		}

		for range int(files) {
			var f File

			_, err := internal.ReadUint8(zr)
			if err != nil {
				return nil, err
			}

			f.Name, err = internal.ReadString(zr)
			if err != nil {
				return nil, err
			}

			f.Size, err = internal.ReadUint64(zr)
			if err != nil {
				return nil, err
			}

			f.Extension, err = internal.ReadString(zr)
			if err != nil {
				return nil, err
			}

			attributes, err := internal.ReadUint32(zr)
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			}

			for range int(attributes) {
				var a Attribute

				code, err := internal.ReadUint32(zr)
				if err != nil {
					return nil, err
				}

				a.Code = FileAttributeType(code)

				a.Value, err = internal.ReadUint32(zr)
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
