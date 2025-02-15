package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
)

// RoomListCode RoomList.
const RoomListCode soul.UInt = 64

type RoomList struct {
	Rooms []Room
}

type Room struct {
	Name     string
	Users    int
	Private  bool
	Owned    bool
	Operated bool
}

func (r *RoomList) Deserialize(reader io.Reader) error {
	soul.ReadUInt(reader)         // size
	code := soul.ReadUInt(reader) // code 64
	if code != RoomListCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", RoomListCode, code))
	}

	// Public room.
	numberOfRooms := soul.ReadUInt(reader)

	// Iterate over the number of rooms and read the room names.
	for i := 0; i < int(numberOfRooms); i++ {
		r.Rooms = append(r.Rooms, Room{
			Name: soul.ReadString(reader),
		})
	}

	for i := range r.Rooms {
		r.Rooms[i].Users = int(soul.ReadUInt(reader))
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

	r.Rooms = append(r.Rooms, ownedPrivateRooms...)

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

	r.Rooms = append(r.Rooms, ownedPrivateRooms...)

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

	r.Rooms = append(r.Rooms, ownedPrivateRooms...)

	return nil
}
