package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/client"
	"github.com/bh90210/soul/peer"
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
		SoulSeekPort:      2242,
		SharedFolders:     100,
		SharedFiles:       1000,
		LogLevel:          zerolog.InfoLevel,
		Timeout:           60 * time.Second,
		LoginTimeout:      10 * time.Second,
		DownloadFolder:    os.TempDir(),
		MaxPeers:          100,
		AcceptChildren:    true,
	}

	// Setup logger.
	log.Logger = log.Level(config.LogLevel)
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Create a new client.
	c, err := client.New(config)
	if err != nil {
		logger.Fatal().Err(err).Msg("new client")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Coonection to the server.
	err = c.Dial(ctx, cancel)
	if err != nil {
		logger.Fatal().Err(err).Msg("dial")
	}

	// We need the state to login search and download.
	state := client.NewState(c)
	err = state.Login(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("login")
	}

	logger.Info().Str("username", config.Username).Msg("logged in")

	// Our search token.
	token := soul.NewToken()
	searchCtx, searchCancel := context.WithCancel(ctx)
	defer searchCancel()

	// Construct the search query.
	query := strings.Join(search, " ")

	// Make the search. It returns a channel we can read the results from.
	results, err := state.Search(searchCtx, query, token)
	if err != nil {
		logger.Fatal().Err(err).Msg("search")
	}

	logger.Info().Any("token", token).Str("query", query).Msg("searching")

	// This is purely for demonstration purposes.
	// We use uilive to show the search results in one line.
	writer := uilive.New()
	writer.Start()

	for {
		// Start reading the results.
		result := <-results
		if result == nil {
			continue
		}

		logger.Info().Str("username", result.Username).Any("result", result).Msg("search result")

		if result.Queue == 0 && result.Size != 0 {
			// var download bool
			// // Filter only lossless files available for download immediately.
			// for _, attribute := range result.File.Attributes {
			// 	if attribute.Code == 4 {
			// 		download = true
			// 		break
			// 	}
			// }

			// if !download {
			// 	continue
			// }

			downloadCtx, downloadCancel := context.WithCancel(ctx)
			defer downloadCancel()

			statusD, errS := state.Download(downloadCtx, result)

			logger.Info().Str("file", result.Name).Str("peer", result.Username).Msg("downloading")
			logger = log.Output(zerolog.ConsoleWriter{Out: writer})

			for {
				select {
				case s := <-statusD:
					logger.Info().Str("file", result.Name).Str("peer", result.Username).Str("status", s).Msg("download status")
					continue

				case e := <-errS:
					if errors.Is(e, peer.ErrComplete) {
						writer.Stop()
						logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
						logger.Info().Str("file", result.Name).Str("peer", result.Username).Msg("download complete")
						downloadCancel()
						return
					}

					logger.Warn().Str("file", result.Name).Str("peer", result.Username).Err(e).Msg("download error")
					return
				}
			}
		}
	}
}
