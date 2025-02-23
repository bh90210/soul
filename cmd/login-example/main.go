package main

import (
	"os"

	"github.com/bh90210/soul/flow"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Level(zerolog.DebugLevel) // TODO: change to info.

	s := flow.Server{
		Config: &flow.Config{
			Username:        "kokomploko123",
			Password:        "sizbty$%YDHFGfg",
			SoulseekAddress: "server.slsknet.org",
			// SoulseekAddress: "localhost", // Local dev.
			SoulseekPort: 2242,
		},
	}

	err := s.Dial()
	if err != nil {
		log.Fatal().Err(err).Msg("dial")
	}

	defer s.Close()

	go func() {
		for {
			_, err := s.NextMessage()
			if err != nil {
				log.Fatal().Err(err).Msg("nextmessage")
			}
		}
	}()

	loginMessage, err := s.Login()
	if err != nil {
		log.Fatal().Err(err).Msg("login")
	}

	log.Debug().Any("login message", loginMessage).Msg("login success")

	go s.Ping()

	select {}
}
