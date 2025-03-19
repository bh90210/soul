package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/client"
	"github.com/gosuri/uilive"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	search := os.Args[1:]
	fmt.Println(search)

	config := &client.Config{
		Username:          "ppoooopiko",
		Password:          "ghfyu5eu6yt",
		OwnPort:           2234,
		OwnObfuscatedPort: 2235,
		SoulSeekAddress:   "server.slsknet.org",
		// SoulSeekAddress: "localhost",
		SoulSeekPort:  2242,
		SharedFolders: 100,
		SharedFiles:   1000,
		// LogLevel:      zerolog.InfoLevel,
		LogLevel:       zerolog.DebugLevel,
		Timeout:        60 * time.Second,
		LoginTimeout:   10 * time.Second,
		DownloadFolder: os.TempDir(),
		MaxPeers:       100,
		AcceptChildren: true,
	}

	// Setup logger.
	log.Logger = log.Level(config.LogLevel)
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	c, err := client.New(config)
	if err != nil {
		logger.Fatal().Err(err).Msg("new client")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = c.Dial(ctx, cancel)
	if err != nil {
		logger.Fatal().Err(err).Msg("dial")
	}

	state := client.NewState(c)
	err = state.Login(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("login")
	}

	logger.Info().Str("username", config.Username).Msg("logged in")

	// select {}
	token := soul.NewToken()
	searchCtx, searchCancel := context.WithCancel(ctx)
	defer searchCancel()

	query := strings.Join(search, " ")

	results, err := state.Search(searchCtx, query, token)
	if err != nil {
		logger.Fatal().Err(err).Msg("search")
	}

	logger.Info().Any("token", token).Str("query", query).Msg("searching")

	writer := uilive.New()

	writer.Start()

	for {
		result := <-results
		if result == nil {
			continue
		}

		logger.Info().Str("username", result.Username).Any("results", result).Msg("search result")

		// 	if result.Queue == 0 && result.Results != nil {
		// 		logger.Info().Int("result", len(result.Results)).Msg("search result")

		// 		if result.Results[0].Size == 0 {
		// 			logger.Warn().Str("file", result.Results[0].Name).Msg("file has no size")
		// 			continue
		// 		}

		// 		downloadCtx, downloadCancel := context.WithCancel(ctx)
		// 		defer downloadCancel()

		// 		statusD, errS := state.Download(downloadCtx, result)

		// 		logger.Info().Str("file", result.Results[0].Name).Str("peer", result.Username).Msg("downloading")
		// 		logger = log.Output(zerolog.ConsoleWriter{Out: writer})

		// 		for {
		// 			select {
		// 			case s := <-statusD:
		// 				logger.Info().Str("file", result.Results[0].Name).Str("peer", result.Username).Str("status", s).Msg("download status")
		// 				continue

		// 			case e := <-errS:
		// 				if errors.Is(e, peer.ErrComplete) {
		// 					writer.Stop()
		// 					logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		// 					logger.Info().Str("file", result.Results[0].Name).Str("peer", result.Username).Msg("download complete")
		// 					downloadCancel()
		// 					return
		// 					// continue
		// 				}

		// 				logger.Error().Str("file", result.Results[0].Name).Str("peer", result.Username).Err(e).Msg("download error")
		// 				return
		// 			}
		// 		}
		// 	}
	}
}
