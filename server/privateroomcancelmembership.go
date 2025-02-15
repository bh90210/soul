package server

import (
	"bytes"

	"github.com/bh90210/soul"
)

const PrivateRoomCancelMembershipCode soul.UInt = 136

type PrivateRoomCancelMembership struct{}

func (p PrivateRoomCancelMembership) Serialize(room string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUInt(buf, PrivateRoomCancelMembershipCode)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, room)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}
