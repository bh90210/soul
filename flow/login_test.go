package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	s := new(Server)
	s.Config = &Config{
		SoulseekAddress: "server.slsknet.org",
		SoulseekPort:    2242,
		Username:        "pipitopapi",
		Password:        "5466854342",
		SharedFolders:   1,
		SharedFiles:     1,
	}

	err := s.Dial()
	assert.NoError(t, err)

	defer s.Close()

	go func() {
		for {
			s.NextMessage()
		}
	}()

	loginMessage, err := s.Login()
	assert.NoError(t, err)
	assert.NotNil(t, loginMessage)
}
