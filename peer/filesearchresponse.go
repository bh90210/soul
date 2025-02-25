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
func (f FileSearchResponse) Serialize(fs FileSearchResponse) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeFileSearchResponse))
	if err != nil {
		return nil, err
	}

	gzw := gzip.NewWriter(buf)

	err = internal.WriteString(gzw, fs.Username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(gzw, uint32(fs.Token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(gzw, uint32(len(fs.Results)))
	if err != nil {
		return nil, err
	}

	err = f.walkWrite(gzw, fs.Results)
	if err != nil {
		return nil, err
	}

	err = internal.WriteBool(gzw, fs.FreeSlot)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(gzw, uint32(fs.AverageSpeed))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(gzw, uint32(fs.Queue))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(gzw, uint32(0))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(gzw, uint32(len(fs.PrivateResults)))
	if err != nil {
		return nil, err
	}

	err = f.walkWrite(gzw, fs.PrivateResults)
	if err != nil {
		return nil, err
	}

	err = gzw.Close()
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (FileSearchResponse) walkWrite(gzw *gzip.Writer, files []File) error {
	err := internal.WriteUint32(gzw, uint32(len(files)))
	if err != nil {
		return err
	}

	for _, file := range files {
		err = internal.WriteUint8(gzw, uint8(1))
		if err != nil {
			return err
		}

		if file.Name == "" {
			return ErrEmptyFileName
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

	return nil
}

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

	gzr, err := gzip.NewReader(reader)
	if err != nil {
		log.Fatal(err)
	}

	defer gzr.Close()

	f.Username, err = internal.ReadString(gzr)
	if err != nil {
		return err
	}

	f.Token, err = internal.ReadUint32ToToken(gzr)
	if err != nil {
		return err
	}

	results, err := internal.ReadUint32(gzr)
	if err != nil {
		return err
	}

	f.Results, err = f.walkRead(results, gzr)
	if err != nil {
	}

	f.FreeSlot, err = internal.ReadBool(gzr)
	if err != nil {
		return err
	}

	f.AverageSpeed, err = internal.ReadUint32ToInt(gzr)
	if err != nil {
		return err
	}

	f.Queue, err = internal.ReadUint32ToInt(gzr)
	if err != nil {
		return err
	}

	_, err = internal.ReadUint32(gzr)
	if err != nil {
		return err
	}

	privateResults, err := internal.ReadUint32(gzr)
	if err != nil {
		return err
	}

	f.PrivateResults, err = f.walkRead(privateResults, gzr)
	if err != nil {
		return err
	}

	return nil
}

func (f *FileSearchResponse) walkRead(numberOfFiles uint32, gzr *gzip.Reader) (files []File, err error) {
	for i := uint32(0); i < numberOfFiles; i++ {
		var file File

		_, err = internal.ReadUint8(gzr)
		if err != nil {
			return
		}

		file.Name, err = internal.ReadString(gzr)
		if err != nil {
			return
		}

		file.Size, err = internal.ReadUint64(gzr)
		if err != nil {
			return
		}

		file.Extension, err = internal.ReadString(gzr)
		if err != nil {
			return
		}

		var attributes uint32
		attributes, err = internal.ReadUint32(gzr)
		if err != nil && !errors.Is(err, io.EOF) {
			return
		}

		for j := uint32(0); j < attributes; j++ {
			attribute := Attribute{}

			var code uint32
			code, err = internal.ReadUint32(gzr)
			if err != nil {
				return
			}

			attribute.Code = FileAttributeType(code)

			attribute.Value, err = internal.ReadUint32(gzr)
			if err != nil && !errors.Is(err, io.EOF) {
				return
			}

			file.Attributes = append(file.Attributes, attribute)
		}

		files = append(files, file)
	}

	return
}
