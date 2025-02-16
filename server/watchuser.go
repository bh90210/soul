package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const WatchUserCode Code = 5

type WatchUser struct {
	Username     string
	Exists       bool
	Status       soul.UserStatus
	AverageSpeed int
	UploadNumber int
	Files        int
	Directories  int
	CountryCode  string
}

// Serialize serializes the WatchUser struct into a byte slice
func (w WatchUser) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	soul.WriteUint32(buf, uint32(WatchUserCode))

	err := soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (w *WatchUser) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 5
	if err != nil {
		return err
	}

	if code != uint32(WatchUserCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", WatchUserCode, code))
	}

	w.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	w.Exists, err = soul.ReadBool(reader)
	if err != nil {
		return err
	}

	if w.Exists {
		status, err := soul.ReadUint32(reader)
		if err != nil {
			return err
		}

		w.Status = soul.UserStatus(status)

		w.AverageSpeed, err = soul.ReadInt(reader)
		if err != nil {
			return err
		}

		w.UploadNumber, err = soul.ReadInt(reader)
		if err != nil {
			return err
		}

		w.Files, err = soul.ReadInt(reader)
		if err != nil {
			return err
		}

		w.Directories, err = soul.ReadInt(reader)
		if err != nil {
			return err
		}
	}

	if w.Status == soul.Online || w.Status == soul.Away {
		w.CountryCode, err = soul.ReadString(reader)
		if err != nil {
			return err
		}
	}

	return nil
}
