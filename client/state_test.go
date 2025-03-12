package client

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	if testing.Short() {
		os.Exit(0)
	}

	t.Parallel()

	user1, err := New()
	assert.NoError(t, err)

	user1.config.SoulSeekAddress = "localhost"
	ctx, cancel := context.WithCancel(context.Background())

	err = user1.Dial(ctx, cancel)
	assert.NoError(t, err)

	state := NewState(user1)

	err = state.Login(ctx)
	assert.NoError(t, err)
}
