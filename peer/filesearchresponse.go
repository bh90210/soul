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

const CodeFileSearchResponse Code = 9

// FileSearchResponse code 9 peer sends this message when it has a file search match.
// The token is taken from original FileSearch, UserSearch or RoomSearch server message.
type FileSearchResponse struct {
	Username       string
	Token          soul.Token
	Results        []File
	FreeSlot       bool
	AverageSpeed   int
	Queue          int // Queue is the length of the queued transfers.
	PrivateResults []File
}

// Serialize accepts a FileSearchResponse and returns a message packed as a byte slice.
func (f *FileSearchResponse) Serialize(fs *FileSearchResponse) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeFileSearchResponse))
	if err != nil {
		return nil, err
	}

	zw := zlib.NewWriter(buf)

	err = internal.WriteString(zw, fs.Username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(zw, uint32(fs.Token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(zw, uint32(len(fs.Results)))
	if err != nil {
		return nil, err
	}

	err = f.walkWrite(zw, fs.Results)
	if err != nil {
		return nil, err
	}

	err = internal.WriteBool(zw, fs.FreeSlot)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(zw, uint32(fs.AverageSpeed))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(zw, uint32(fs.Queue))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(zw, uint32(0))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(zw, uint32(len(fs.PrivateResults)))
	if err != nil {
		return nil, err
	}

	err = f.walkWrite(zw, fs.PrivateResults)
	if err != nil {
		return nil, err
	}

	err = zw.Close()
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (FileSearchResponse) walkWrite(zw *zlib.Writer, files []File) error {
	for _, file := range files {
		err := internal.WriteUint8(zw, uint8(1))
		if err != nil {
			return err
		}

		if file.Name == "" {
			return ErrEmptyFileName
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

	return nil
}

// Deserialize populates a FileSearchResponse with the data in the provided reader.
func (f *FileSearchResponse) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 9
	if err != nil {
		return err
	}

	if code != uint32(CodeFileSearchResponse) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeFileSearchResponse, code))
	}

	zr, err := zlib.NewReader(reader)
	if err != nil {
		return err
	}

	defer zr.Close()

	f.Username, err = internal.ReadString(zr)
	if err != nil {
		return err
	}

	f.Token, err = internal.ReadUint32ToToken(zr)
	if err != nil {
		return err
	}

	results, err := internal.ReadUint32(zr)
	if err != nil {
		return err
	}

	f.Results, err = f.walkRead(results, zr)
	if err != nil {
		return err
	}

	f.FreeSlot, err = internal.ReadBool(zr)
	if err != nil {
		return err
	}

	f.AverageSpeed, err = internal.ReadUint32ToInt(zr)
	if err != nil {
		return err
	}

	f.Queue, err = internal.ReadUint32ToInt(zr)
	if err != nil {
		return err
	}

	_, err = internal.ReadUint32(zr)
	if err != nil {
		return err
	}

	privateResults, err := internal.ReadUint32(zr)
	if err != nil {
		return err
	}

	f.PrivateResults, err = f.walkRead(privateResults, zr)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return err
}

func (f *FileSearchResponse) walkRead(numberOfFiles uint32, zr io.ReadCloser) (files []File, err error) {
	for i := uint32(0); i < numberOfFiles; i++ {
		var file File

		_, err = internal.ReadUint8(zr)
		if err != nil {
			return
		}

		file.Name, err = internal.ReadString(zr)
		if err != nil {
			return
		}

		file.Size, err = internal.ReadUint64(zr)
		if err != nil {
			return
		}

		file.Extension, err = internal.ReadString(zr)
		if err != nil {
			return
		}

		var attributes uint32
		attributes, err = internal.ReadUint32(zr)
		if err != nil && !errors.Is(err, io.EOF) {
			return
		}
		for j := uint32(0); j < attributes; j++ {
			attribute := Attribute{}

			var code uint32
			code, err = internal.ReadUint32(zr)
			if err != nil {
				return
			}

			attribute.Code = FileAttributeType(code)

			attribute.Value, err = internal.ReadUint32(zr)
			if err != nil && !errors.Is(err, io.EOF) {
				return
			}

			file.Attributes = append(file.Attributes, attribute)
		}

		files = append(files, file)
	}

	return
}
