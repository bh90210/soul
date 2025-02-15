package server

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/bh90210/soul"
)

const GetUserStatsCode soul.UInt = 36

type GetUserStats struct {
	Username    string
	Speed       int
	Uploads     int
	Files       int
	Directories int
}

func (g GetUserStats) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, GetUserStatsCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (g *GetUserStats) Deserialize(reader *bytes.Reader) error {
	_, err := soul.ReadUInt(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUInt(reader) // code 36
	if err != nil {
		return err
	}

	if code != GetUserStatsCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", GetUserStatsCode, code))
	}

	g.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	g.Speed, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	g.Uploads, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	g.Files, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	g.Directories, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	return nil
}
