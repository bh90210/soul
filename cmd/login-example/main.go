package main

import (
	"os"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/flow"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Init a server instance.
	client := new(flow.Client)

	// Setup the server configuration.
	username, _ := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz", 12)
	password, _ := gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 12)

	client.Config = &flow.Config{
		Username:        username,
		Password:        password,
		SoulseekAddress: "server.slsknet.org",
		// SoulseekAddress: "localhost", // Local dev.
		SoulseekPort:  2242,
		SharedFolders: 1,
		SharedFiles:   1,
		LogLevel:      zerolog.DebugLevel,
	}

	// Setup logger.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Level(client.Config.LogLevel)

	// Connect to the server.
	err := client.Dial()
	if err != nil {
		log.Fatal().Err(err).Msg("dial")
	}

	defer client.Close()

	// Start listening for messages.
	go func() {
		for {
			_, err := client.NextMessage()
			if err != nil {
				log.Fatal().Err(err).Msg("nextmessage")
			}
		}
	}()

	// Login to the server.
	loginMessage, err := client.Login()
	if err != nil {
		log.Fatal().Err(err).Msg("login")
	}

	log.Debug().Any("login message", loginMessage.Login).Msg("login success")

	// When connected, start pinging the server.
	go client.Ping()

	time.Sleep(2 * time.Second)

	token := soul.NewToken()

	err = client.GlobalSearch(token, "bob marley kaya")
	if err != nil {
		log.Fatal().Err(err).Msg("global search")
	}

	go func() {
		for {
			res := client.PollSearchResults(token)
			log.Debug().Any("search results", res).Msg("poll search results")
			time.Sleep(5000 * time.Millisecond)
		}
	}()

	select {}
}
