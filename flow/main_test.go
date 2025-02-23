package flow

import (
	"log"
	"testing"
)

var s *Client

func TestMain(m *testing.M) {
	s = new(Client)
	s.Config = &Config{
		SoulseekAddress: "server.slsknet.org",
		SoulseekPort:    2242,
		Username:        "pipitopapi",
		Password:        "5466854342",
		SharedFolders:   1,
		SharedFiles:     10,
	}

	err := s.Dial()
	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	go func() {
		for {
			s.NextMessage()
		}
	}()

	m.Run()
}
