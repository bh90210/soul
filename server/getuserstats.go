package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const GetUserStatsCode soul.ServerCode = 36

type GetUserStats struct {
	Username    string
	Speed       int
	Uploads     int
	Files       int
	Directories int
}

func (g GetUserStats) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(GetUserStatsCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (g *GetUserStats) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 36
	if err != nil {
		return err
	}

	if code != uint32(GetUserStatsCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", GetUserStatsCode, code))
	}

	g.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	g.Speed, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	g.Uploads, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	g.Files, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	g.Directories, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
