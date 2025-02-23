package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	loginMessage, err := s.Login()
	assert.NoError(t, err)
	assert.NotNil(t, loginMessage)
}
