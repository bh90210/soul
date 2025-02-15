package server

import (
	"encoding/hex"
	"testing"
)

func TestLoginMessage(t *testing.T) {
	l := new(Login)
	r, err := l.Serialize("username", "password")
	if err != nil {
		t.Fail()
	}

	having := hex.EncodeToString(r)
	expecting := "480000000100000008000000757365726e616d650800000070617373776f7264a000000020000000643531633961376539333533373436613630323066393630326434353239323901000000"
	if having != expecting {
		t.Fail()
	}
}
