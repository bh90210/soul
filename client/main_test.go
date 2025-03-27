package client

import (
	"flag"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()

	if testing.Short() {
		return
	}

	m.Run()
}
