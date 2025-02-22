package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

const JoinRoomCode soul.ServerCode = 14

type JoinRoom struct {
	Room  string
	Users []User

	Private   bool
	Owner     string
	Operators []string
}

type User struct {
	Username string
	Status   UserStatus

	AverageSpeed int
	UploadNumber int
	Files        int
	Directories  int

	FreeSlots int

	CountryCode string
}

func (j JoinRoom) Serialize(room string, private bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(JoinRoomCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	if private {
		soul.WriteUint32(buf, 1)
	} else {
		soul.WriteUint32(buf, 0)
	}

	return soul.Pack(buf.Bytes())
}

func (j *JoinRoom) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 1
	if err != nil {
		return err
	}

	if code != uint32(JoinRoomCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", JoinRoomCode, code))
	}

	j.Room, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	usersInRoom, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(usersInRoom); i++ {
		var u User

		u.Username, err = soul.ReadString(reader)
		if err != nil {
			return err
		}

		j.Users = append(j.Users, u)
	}

	statuses, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(statuses); i++ {
		status, err := soul.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].Status = UserStatus(status)
	}

	stats, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(stats); i++ {
		speed, err := soul.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].AverageSpeed = int(speed)

		upload, err := soul.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].UploadNumber = int(upload)

		_, err = soul.ReadUint32(reader)
		if err != nil {
			return err
		}

		files, err := soul.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].Files = int(files)

		directories, err := soul.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].Directories = int(directories)
	}

	slots, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(slots); i++ {
		freeSlots, err := soul.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].FreeSlots = int(freeSlots)
	}

	countries, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(countries); i++ {
		countryCode, err := soul.ReadString(reader)
		if err != nil {
			return err
		}

		j.Users[i].CountryCode = countryCode
	}

	j.Owner, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	operators, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(operators); i++ {
		operator, err := soul.ReadString(reader)
		if err != nil {
			return err
		}

		j.Operators = append(j.Operators, operator)
	}

	return nil
}
