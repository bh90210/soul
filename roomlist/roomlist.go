package roomlist

import (
	"io"

	"github.com/bh90210/soul"
)

// Code RoomList.
const Code soul.UInt = 64

type Response struct {
	Rooms []Room
}

type Room struct {
	Name     string
	Users    int
	Private  bool
	Owned    bool
	Operated bool
}

func Deserialize(reader io.Reader) *Response {
	soul.ReadUInt(reader) // size
	soul.ReadUInt(reader) // code 64

	// Public room.
	numberOfRooms := soul.ReadUInt(reader)

	// Iterate over the number of rooms and read the room names.
	rooms := make([]Room, 0)
	for i := 0; i < int(numberOfRooms); i++ {
		rooms = append(rooms, Room{
			Name: soul.ReadString(reader),
		})
	}

	for i := range rooms {
		rooms[i].Users = int(soul.ReadUInt(reader))
	}

	// Owned private rooms.
	numberOfPrivateRooms := soul.ReadUInt(reader)

	ownedPrivateRooms := make([]Room, 0)
	for i := 0; i < int(numberOfPrivateRooms); i++ {
		ownedPrivateRooms = append(ownedPrivateRooms, Room{
			Name:    soul.ReadString(reader),
			Private: true,
			Owned:   true,
		})
	}

	for i := range ownedPrivateRooms {
		ownedPrivateRooms[i].Users = int(soul.ReadUInt(reader))
	}

	rooms = append(rooms, ownedPrivateRooms...)

	// Not owned private rooms.
	numberOFNotOwnedPrivateRooms := soul.ReadUInt(reader)

	notOwnedPrivateRooms := make([]Room, 0)
	for i := 0; i < int(numberOFNotOwnedPrivateRooms); i++ {
		notOwnedPrivateRooms = append(notOwnedPrivateRooms, Room{
			Name:    soul.ReadString(reader),
			Private: true,
		})
	}

	for i := range notOwnedPrivateRooms {
		notOwnedPrivateRooms[i].Users = int(soul.ReadUInt(reader))
	}

	rooms = append(rooms, ownedPrivateRooms...)

	// Operated private rooms.
	numberOfOperatedPrivateRooms := soul.ReadUInt(reader)

	operatedPrivateRooms := make([]Room, 0)
	for i := 0; i < int(numberOfOperatedPrivateRooms); i++ {
		operatedPrivateRooms = append(operatedPrivateRooms, Room{
			Name:     soul.ReadString(reader),
			Private:  true,
			Operated: true,
		})
	}

	for i := range operatedPrivateRooms {
		operatedPrivateRooms[i].Users = int(soul.ReadUInt(reader))
	}

	rooms = append(rooms, ownedPrivateRooms...)

	return &Response{rooms}
}
