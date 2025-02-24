package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	loginMessage, err := user1.Login()
	assert.NoError(t, err)
	assert.NotNil(t, loginMessage)
}
