package server

import (
	"encoding/hex"
	"testing"
)

func TestWritePort(t *testing.T) {
	swp := new(SetListenPort)
	v, _ := swp.Serialize(2234)
	having := hex.EncodeToString(v)
	expecting := "0400000002000000"
	if having != expecting {
		t.Fail()
	}
}
