package server

import (
	"bytes"
	"encoding/binary"

	"github.com/bh90210/soul"
)

const JoinRoomCode soul.UInt = 14

type JoinRoom struct {
	Room  string
	Users []User

	Owner     string
	Operators []string
}

type User struct {
	Username string
	Status   soul.UserStatusCode

	AverageSpeed int
	UploadNumber int
	Files        int
	Directories  int

	FreeSlots int

	CountryCode string
}

func (j JoinRoom) Serialize(room string, private bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, JoinRoomCode)
	if err != nil {
		return nil, err
	}

	r, err := soul.NewString(room)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, r)
	if err != nil {
		return nil, err
	}

	if private {
		soul.WriteUInt(buf, 1)
	} else {
		soul.WriteUInt(buf, 0)
	}

	return soul.Pack(buf.Bytes())
}
