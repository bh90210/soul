package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/client"
	"github.com/bh90210/soul/peer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	search := os.Args[1:]
	// flac := flag.Bool("flac", true, "download flac only")
	// flag.Parse()

	config := &client.Config{
		Username: "suigfiuggg",
		Password: "sfdgdhu56i8456",
		OwnPort:  2234,
		// SoulSeekAddress: "server.slsknet.org",
		SoulSeekAddress: "localhost",
		SoulSeekPort:    2242,
		SharedFolders:   1,
		SharedFiles:     1,
		// LogLevel:      zerolog.InfoLevel,
		LogLevel:       zerolog.DebugLevel,
		Timeout:        60 * time.Second,
		LoginTimeout:   3 * time.Second,
		DownloadFolder: os.TempDir(),
		MaxPeers:       1,
	}

	// Setup logger.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.Level(config.LogLevel)

	c, err := client.New(config)
	if err != nil {
		log.Fatal().Err(err).Msg("new client")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = c.Dial(ctx, cancel)
	if err != nil {
		log.Fatal().Err(err).Msg("dial")
	}

	state := client.NewState(c)
	err = state.Login(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("login")
	}

	log.Info().Str("username", config.Username).Msg("logged in")

	token := soul.NewToken()
	searchCtx, searchCancel := context.WithCancel(ctx)
	defer searchCancel()

	query := strings.Join(search, " ")

	results, err := state.Search(searchCtx, query, token)
	if err != nil {
		log.Fatal().Err(err).Msg("search")
	}

	log.Info().Any("token", token).Str("query", query).Msg("searching")

	for {
		result := <-results
		if result == nil {
			continue
		}

		log.Info().Int("result", len(result.Results)).Msg("search result")

		if result.Queue == 0 && result.Results != nil {
			if result.Results[0].Size == 0 {
				log.Warn().Str("file", result.Results[0].Name).Msg("file has no size")
				continue
			}

			downloadCtx, downloadCancel := context.WithCancel(ctx)
			defer downloadCancel()

			statusD, errS := state.Download(downloadCtx, client.Download{
				Username: result.Username,
				Token:    token,
				File:     &result.Results[0],
			})

			log.Info().Str("file", result.Results[0].Name).Str("peer", result.Username).Msg("downloading")

			for {
				select {
				case s := <-statusD:
					log.Info().Str("file", result.Results[0].Name).Str("peer", result.Username).Str("status", s).Msg("download status")
					continue

				case e := <-errS:
					if errors.Is(e, peer.ErrComplete) {
						log.Info().Str("file", result.Results[0].Name).Str("peer", result.Username).Msg("download complete")
						return
					}

					log.Error().Str("file", result.Results[0].Name).Str("peer", result.Username).Err(e).Msg("download error")
					return
				}
			}
		}
	}
}
