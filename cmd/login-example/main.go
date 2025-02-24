package main

import (
	"os"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/flow"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Init a server instance.
	client := new(flow.Client)

	// Setup the server configuration.
	client.Config = &flow.Config{
		Username:        "kokomploko123",
		Password:        "sizbty$%YDHFGfg",
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

	var token soul.Token
	token.Gen()

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
