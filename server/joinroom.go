package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodeJoinRoom Code = 14

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

func (j *JoinRoom) Serialize(message *JoinRoom) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeJoinRoom))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Room)
	if err != nil {
		return nil, err
	}

	if message.Private {
		internal.WriteUint32(buf, 1)
	} else {
		internal.WriteUint32(buf, 0)
	}

	return internal.Pack(buf.Bytes())
}

func (j *JoinRoom) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 1
	if err != nil {
		return err
	}

	if code != uint32(CodeJoinRoom) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeJoinRoom, code))
	}

	j.Room, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	usersInRoom, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for range int(usersInRoom) {
		var u User

		u.Username, err = internal.ReadString(reader)
		if err != nil {
			return err
		}

		j.Users = append(j.Users, u)
	}

	statuses, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(statuses); i++ {
		status, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].Status = UserStatus(status)
	}

	stats, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := range int(stats) {
		speed, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].AverageSpeed = int(speed)

		upload, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].UploadNumber = int(upload)

		_, err = internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		files, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].Files = int(files)

		directories, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].Directories = int(directories)
	}

	slots, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := range int(slots) {
		freeSlots, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		j.Users[i].FreeSlots = int(freeSlots)
	}

	countries, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := range int(countries) {
		countryCode, err := internal.ReadString(reader)
		if err != nil {
			return err
		}

		j.Users[i].CountryCode = countryCode
	}

	j.Owner, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	operators, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for range int(operators) {
		operator, err := internal.ReadString(reader)
		if err != nil {
			return err
		}

		j.Operators = append(j.Operators, operator)
	}

	return nil
}
