package peer

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const SharedFileListResponseCode soul.PeerCode = 5

type SharedFileListResponse struct {
	Directories        []Directory
	PrivateDirectories []Directory
}

type Directory struct {
	Name  string
	Files []File
}

type File struct {
	Name       string
	Size       uint64
	Extension  string
	Attributes []Attribute
}

type Attribute struct {
	Code  FileAttributeType
	Value int
}

func (s SharedFileListResponse) Serialize(directories, privateDirectories []Directory) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(SharedFileListResponseCode))
	if err != nil {
		return nil, err
	}

	gzw := gzip.NewWriter(buf)

	err = internal.WriteUint32(gzw, uint32(len(directories)))
	if err != nil {
		return nil, err
	}

	err = s.walkWrite(directories, gzw)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(gzw, 0)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(gzw, uint32(len(privateDirectories)))
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

var ErrEmptyFileDirectory = errors.New("directory is empty")

var ErrEmptyDirectoryName = errors.New("directory name is empty")

var ErrEmptyFileName = errors.New("file name is empty")

var ErrSizeZero = errors.New("file size is zero")

var ErrEmptyFileExtension = errors.New("file extension is empty")

func (s SharedFileListResponse) walkWrite(directories []Directory, gzw *gzip.Writer) error {
	for _, directory := range directories {
		if directory.Name == "" {
			return ErrEmptyDirectoryName
		}

		if len(directory.Files) == 0 {
			return ErrEmptyFileDirectory
		}

		// err := internal.WriteString(gzw, directory.Name)
	}

	return nil
}

func (s *SharedFileListResponse) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 5
	if err != nil {
		return err
	}

	if code != uint32(SharedFileListResponseCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", SharedFileListResponseCode, code))
	}

	gzr, err := gzip.NewReader(reader)
	if err != nil {
		log.Fatal(err)
	}

	directories, err := internal.ReadUint32(gzr)
	if err != nil {
		return err
	}

	dirs, err := s.walkRead(directories, gzr)
	if err != nil {
		return err
	}

	s.Directories = append(s.Directories, dirs...)

	_, err = internal.ReadUint32(gzr)
	if err != nil {
		return err
	}

	privateDirectories, err := internal.ReadUint32(gzr)
	if err != nil {
		return err
	}

	privateDirs, err := s.walkRead(privateDirectories, gzr)
	if err != nil {
		return err
	}

	s.PrivateDirectories = append(s.PrivateDirectories, privateDirs...)

	return nil
}

func (s SharedFileListResponse) walkRead(numberOfDirectories uint32, gzr *gzip.Reader) ([]Directory, error) {
	var directories []Directory

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
			if err != nil {
				return nil, err
			}

			for k := 0; k < int(attributes); k++ {
				var a Attribute

				code, err := internal.ReadUint32(gzr)
				if err != nil {
					return nil, err
				}

				a.Code = FileAttributeType(code)

				a.Value, err = internal.ReadUint32ToInt(gzr)
				if err != nil {
					return nil, err
				}

				f.Attributes = append(f.Attributes, a)
			}

			directory.Files = append(directory.Files, f)
		}

		directories = append(directories, directory)
	}

	return directories, nil
}
