package client

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestDownload(t *testing.T) {
	if testing.Short() {
		return
	}

	// User 1 Login.
	user1, err := New()
	assert.NoError(t, err)

	user1.config.SoulSeekAddress = "localhost"
	user1.config.LogLevel = zerolog.Disabled
	ctx, cancel := context.WithCancel(context.Background())

	err = user1.Dial(ctx, cancel)
	assert.NoError(t, err)

	state1 := NewState(user1)

	err = state1.Login(ctx)
	assert.NoError(t, err)

	// User 2 Login.
	user2, err := New()
	assert.NoError(t, err)

	user2.config.SoulSeekAddress = "localhost"
	user2.config.LogLevel = zerolog.Disabled
	user2.config.OwnPort = 2236
	user2.config.OwnPortObfuscated = 2237
	ctx, cancel = context.WithCancel(context.Background())

	err = user2.Dial(ctx, cancel)
	assert.NoError(t, err)

	state2 := NewState(user2)

	err = state2.Login(ctx)
	assert.NoError(t, err)
}
