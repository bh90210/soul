package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeWatchUser Code = 5

type WatchUser struct {
	Username     string
	Exists       bool
	Status       UserStatus
	AverageSpeed int
	UploadNumber int
	Files        int
	Directories  int
	CountryCode  string
}

// Serialize serializes the WatchUser struct into a byte slice
func (w *WatchUser) Serialize(message *WatchUser) ([]byte, error) {
	buf := new(bytes.Buffer)
	internal.WriteUint32(buf, uint32(CodeWatchUser))

	err := internal.WriteString(buf, message.Username)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (w *WatchUser) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 5
	if err != nil {
		return err
	}

	if code != uint32(CodeWatchUser) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeWatchUser, code))
	}

	w.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	w.Exists, err = internal.ReadBool(reader)
	if err != nil {
		return err
	}

	if w.Exists {
		status, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		w.Status = UserStatus(status)

		w.AverageSpeed, err = internal.ReadUint32ToInt(reader)
		if err != nil {
			return err
		}

		w.UploadNumber, err = internal.ReadUint32ToInt(reader)
		if err != nil {
			return err
		}

		w.Files, err = internal.ReadUint32ToInt(reader)
		if err != nil {
			return err
		}

		w.Directories, err = internal.ReadUint32ToInt(reader)
		if err != nil {
			return err
		}
	}

	if w.Status == StatusOnline || w.Status == StatusAway {
		w.CountryCode, err = internal.ReadString(reader)
		if err != nil {
			return err
		}
	}

	return nil
}
