package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// RoomListCode RoomList.
const RoomListCode soul.CodeServer = 64

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

func (r *RoomList) Deserialize(reader io.Reader) (err error) {
	_, err = internal.ReadUint32(reader) // size
	if err != nil {
		return
	}

	code, err := internal.ReadUint32(reader) // code 64
	if err != nil {
		return
	}

	if code != uint32(RoomListCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", RoomListCode, code))
	}

	// Public room.
	numberOfRooms, err := internal.ReadUint32(reader)
	if err != nil {
		return
	}

	// Iterate over the number of rooms and read the room names.
	for i := 0; i < int(numberOfRooms); i++ {
		var name string
		name, err = internal.ReadString(reader)
		if err != nil {
			return
		}

		r.Rooms = append(r.Rooms, Room{
			Name: name,
		})
	}

	for i := range r.Rooms {
		var users uint32
		users, err = internal.ReadUint32(reader)
		if err != nil {
			return
		}

		r.Rooms[i].Users = int(users)
	}

	// Owned private rooms.
	numberOfPrivateRooms, err := internal.ReadUint32(reader)
	if err != nil {
		return
	}

	var ownedPrivateRooms []Room
	for i := 0; i < int(numberOfPrivateRooms); i++ {
		var name string
		name, err = internal.ReadString(reader)
		if err != nil {
			return
		}

		ownedPrivateRooms = append(ownedPrivateRooms, Room{
			Name:    name,
			Private: true,
			Owned:   true,
		})
	}

	for i := range ownedPrivateRooms {
		var no uint32
		no, err = internal.ReadUint32(reader)
		if err != nil {
			return
		}

		ownedPrivateRooms[i].Users = int(no)
	}

	r.Rooms = append(r.Rooms, ownedPrivateRooms...)

	// Not owned private rooms.
	no, err := internal.ReadUint32(reader)
	if err != nil {
		return
	}

	numberOFNotOwnedPrivateRooms := no

	var notOwnedPrivateRooms []Room
	for i := 0; i < int(numberOFNotOwnedPrivateRooms); i++ {
		var name string
		name, err = internal.ReadString(reader)
		if err != nil {
			return
		}

		notOwnedPrivateRooms = append(notOwnedPrivateRooms, Room{
			Name:    name,
			Private: true,
		})
	}

	for i := range notOwnedPrivateRooms {
		no, err = internal.ReadUint32(reader)
		if err != nil {
			return
		}

		notOwnedPrivateRooms[i].Users = int(no)
	}

	r.Rooms = append(r.Rooms, ownedPrivateRooms...)

	// Operated private rooms.
	no, err = internal.ReadUint32(reader)
	if err != nil {
		return
	}

	numberOfOperatedPrivateRooms := no

	var operatedPrivateRooms []Room
	for i := 0; i < int(numberOfOperatedPrivateRooms); i++ {
		var name string
		name, err = internal.ReadString(reader)
		if err != nil {
			return
		}

		operatedPrivateRooms = append(operatedPrivateRooms, Room{
			Name:     name,
			Private:  true,
			Operated: true,
		})
	}

	for i := range operatedPrivateRooms {
		no, err = internal.ReadUint32(reader)
		if err != nil && !errors.Is(err, io.EOF) {
			return
		}

		operatedPrivateRooms[i].Users = int(no)
	}

	r.Rooms = append(r.Rooms, ownedPrivateRooms...)

	return
}
