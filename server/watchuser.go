package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const WatchUserCode soul.UInt = 5

type WatchUser struct {
	Username     string
	Exists       bool
	Status       soul.UserStatusCode
	AverageSpeed int
	UploadNumber int
	Files        int
	Directories  int
	CountryCode  string
}

// Serialize serializes the WatchUser struct into a byte slice
func (w WatchUser) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	soul.WriteUInt(buf, WatchUserCode)

	err := binary.Write(buf, binary.LittleEndian, username)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (w *WatchUser) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 5
	if err != nil {
		return err
	}

	if code != WatchUserCode {
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
		status, err := soul.ReadUInt(reader)
		if err != nil {
			return err
		}

		w.Status = soul.UserStatusCode(status)

		speed, err := soul.ReadUInt(reader)
		if err != nil {
			return err
		}

		w.AverageSpeed = int(speed)

		number, err := soul.ReadUInt(reader)
		if err != nil {
			return err
		}

		w.UploadNumber = int(number)

		files, err := soul.ReadUInt(reader)
		if err != nil {
			return err
		}

		w.Files = int(files)

		directories, err := soul.ReadUInt(reader)
		if err != nil {
			return err
		}

		w.Directories = int(directories)
	}

	if w.Status == soul.Online || w.Status == soul.Away {
		code, err := soul.ReadString(reader)
		if err != nil {
			return err
		}

		w.CountryCode = code
	}

	return nil
}
