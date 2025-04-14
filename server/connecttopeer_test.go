package server

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
	"github.com/bh90210/soul/peer"
	"github.com/stretchr/testify/assert"
)

func TestConnectToPeer(t *testing.T) {
	t.Parallel()

	token := soul.NewToken()

	ctp := new(ConnectToPeer)
	ctp.Token = token
	ctp.Username = "test"
	ctp.Type = peer.ConnectionType

	message, err := ctp.Serialize(ctp)
	assert.NoError(t, err)
	assert.Equal(t, 25, len(message))

	buf := new(bytes.Buffer)
	err = internal.WriteUint32(buf, uint32(CodeConnectToPeer))
	assert.NoError(t, err)
	err = internal.WriteString(buf, ctp.Username)
	assert.NoError(t, err)
	err = internal.WriteString(buf, string(ctp.Type))
	assert.NoError(t, err)
	ip := make(net.IP, 4)
	ip[0] = 127
	ip[1] = 0
	ip[2] = 0
	ip[3] = 1
	err = internal.WriteUint32(buf, binary.BigEndian.Uint32(ip))
	assert.NoError(t, err)
	err = internal.WriteUint32(buf, uint32(ctp.Port))
	assert.NoError(t, err)
	err = internal.WriteUint32(buf, uint32(ctp.Token))
	assert.NoError(t, err)
	err = internal.WriteBool(buf, ctp.Privileged)
	assert.NoError(t, err)
	err = internal.WriteUint32(buf, uint32(ctp.ObfuscatedPort))
	assert.NoError(t, err)
	b, err := internal.Pack(buf.Bytes())
	assert.NoError(t, err)

	ctp = new(ConnectToPeer)
	err = ctp.Deserialize(bytes.NewReader(b))
	assert.NoError(t, err)
	assert.Equal(t, token, ctp.Token)
	assert.Equal(t, "test", ctp.Username)
	assert.Equal(t, peer.ConnectionType, ctp.Type)
	assert.Equal(t, ip.String(), ctp.IP.String())
	assert.Equal(t, 0, ctp.Port)
	assert.Equal(t, false, ctp.Privileged)
	assert.Equal(t, 0, ctp.ObfuscatedPort)
	assert.Equal(t, token, ctp.Token)
}
