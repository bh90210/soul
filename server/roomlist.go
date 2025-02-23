package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// RoomListCode RoomList.
const RoomListCode soul.ServerCode = 64

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
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 64
	if err != nil {
		return err
	}

	if code != uint32(RoomListCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", RoomListCode, code))
	}

	// Public room.
	numberOfRooms, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	// Iterate over the number of rooms and read the room names.
	for i := 0; i < int(numberOfRooms); i++ {
		name, err := internal.ReadString(reader)
		if err != nil {
			return err
		}

		r.Rooms = append(r.Rooms, Room{
			Name: name,
		})
	}

	for i := range r.Rooms {
		users, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		r.Rooms[i].Users = int(users)
	}

	// Owned private rooms.
	numberOfPrivateRooms, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	var ownedPrivateRooms []Room
	for i := 0; i < int(numberOfPrivateRooms); i++ {
		name, err := internal.ReadString(reader)
		if err != nil {
			return err
		}

		ownedPrivateRooms = append(ownedPrivateRooms, Room{
			Name:    name,
			Private: true,
			Owned:   true,
		})
	}

	for i := range ownedPrivateRooms {
		no, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		ownedPrivateRooms[i].Users = int(no)
	}

	r.Rooms = append(r.Rooms, ownedPrivateRooms...)

	// Not owned private rooms.
	no, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	numberOFNotOwnedPrivateRooms := no

	var notOwnedPrivateRooms []Room
	for i := 0; i < int(numberOFNotOwnedPrivateRooms); i++ {
		name, err := internal.ReadString(reader)
		if err != nil {
			return err
		}

		notOwnedPrivateRooms = append(notOwnedPrivateRooms, Room{
			Name:    name,
			Private: true,
		})
	}

	for i := range notOwnedPrivateRooms {
		no, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		notOwnedPrivateRooms[i].Users = int(no)
	}

	r.Rooms = append(r.Rooms, ownedPrivateRooms...)

	// Operated private rooms.
	no, err = internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	numberOfOperatedPrivateRooms := no

	var operatedPrivateRooms []Room
	for i := 0; i < int(numberOfOperatedPrivateRooms); i++ {
		name, err := internal.ReadString(reader)
		if err != nil {
			return err
		}

		operatedPrivateRooms = append(operatedPrivateRooms, Room{
			Name:     name,
			Private:  true,
			Operated: true,
		})
	}

	for i := range operatedPrivateRooms {
		no, err := internal.ReadUint32(reader)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		operatedPrivateRooms[i].Users = int(no)
	}

	r.Rooms = append(r.Rooms, ownedPrivateRooms...)

	return io.EOF
}
