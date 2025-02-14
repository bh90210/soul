package login

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"

	soul "github.com/bh90210/soul"
)

const (
	Code soul.UInt = 1
)

type Response struct {
	*Success
	*Failure
}

func (r Response) OK() bool {
	if r.Failure != nil && r.Success == nil {
		return false
	} else if r.Failure == nil && r.Success != nil {
		return true
	}

	return false
}

type Success struct {
	Greet string
	IP    net.IP
	Sum   string
}

type Failure struct {
	Reason string
}

func Read(reader io.Reader) Response {
	soul.ReadUInt(reader) // size
	soul.ReadUInt(reader) // code 1
	success := soul.ReadBool(reader)
	if success {
		return readSuccess(reader)
	}
	return readFailure(reader)
}

func Write(username string, password string) []byte {
	buf := new(bytes.Buffer)
	soul.WriteUInt(buf, Code)
	binary.Write(buf, binary.LittleEndian, soul.NewString(username))
	binary.Write(buf, binary.LittleEndian, soul.NewString(password))
	binary.Write(buf, binary.LittleEndian, soul.MajorVersion)
	binary.Write(buf, binary.LittleEndian, sum(username, password))
	binary.Write(buf, binary.LittleEndian, soul.MinorVersion)
	return soul.Pack(buf.Bytes())
}

func readIP(val soul.UInt) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, uint32(val))
	return ip
}

func readSuccess(reader io.Reader) Response {
	greet := soul.ReadString(reader)
	ip := readIP(soul.ReadUInt(reader))
	sum := soul.ReadString(reader)
	return Response{Success: &Success{greet, ip, sum}}
}

func readFailure(reader io.Reader) Response {
	reason := soul.ReadString(reader)
	return Response{Failure: &Failure{reason}}
}

func sum(username string, password string) soul.String {
	sum := md5.Sum([]byte(username + password))
	return soul.NewString(hex.EncodeToString(sum[:]))
}
