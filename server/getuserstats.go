package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
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
	err := internal.WriteUint32(buf, uint32(GetUserStatsCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (g *GetUserStats) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 36
	if err != nil {
		return err
	}

	if code != uint32(GetUserStatsCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", GetUserStatsCode, code))
	}

	g.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	g.Speed, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	g.Uploads, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	g.Files, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	g.Directories, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
