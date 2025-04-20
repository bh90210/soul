// Description: This example demonstrates how to search for a file and download it.
// The search is done by sending a search query to the server and reading the results.
package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/client"
	"github.com/bh90210/soul/peer"
	"github.com/gosuri/uilive"
	"github.com/ipsn/go-adorable"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	search := os.Args[1:]

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal().Err(err).Msg("load .env")
	}

	config := &client.Config{
		Username: "te",
		// Username:          gonanoid.MustGenerate("soulseek", 7),
		Password: "password",
		// Password:          gonanoid.MustGenerate("0123456789qwertyuiop", 10),
		OwnPort:           2235,
		OwnPortObfuscated: 2236,
		SoulSeekAddress:   "localhost",
		// SoulSeekAddress: "server.slsknet.org",
		SoulSeekPort:       2242,
		SharedFolders:      100,
		SharedFiles:        1000,
		LogLevel:           zerolog.DebugLevel,
		Timeout:            60 * time.Second,
		LoginTimeout:       10 * time.Second,
		DownloadFolder:     os.TempDir(),
		MaxFileConnections: 5,
		MaxPeers:           100,
		AcceptChildren:     true,
		Picture:            adorable.Random(),
		Library:            os.Getenv("SOUL_LIBRARY"),
		Description:        "soul client",
	}

	// TODO: delete this.
	config.LoginTimeout = time.Second

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

	// Listen for incoming search requests.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case request := <-state.Incoming:
				ctx, _ := context.WithTimeout(ctx, 5*time.Minute) // TODO: timeout from new config field?
				f, err := os.OpenFile(os.Getenv("SOUL_TESTFILE"), os.O_RDONLY, 0644)
				if err != nil {
					logger.Error().Err(err).Msg("read file")
					continue
				}

				info, err := f.Stat()
				if err != nil {
					logger.Error().Err(err).Msg("stat file")
					continue
				}

				files := []*client.File{
					{
						Username: request.Username,
						Token:    request.Token,
						Queue:    0,
						File: &peer.File{
							Name:       f.Name(),
							Size:       uint64(info.Size()),
							Extension:  filepath.Ext(f.Name()),
							Attributes: []peer.Attribute{{Code: 1}},
						},
					},
				}

				err = state.Respond(ctx, files)
				if err != nil {
					logger.Error().Err(err).Msg("respond")
				}
			}
		}
	}()

	// TODO: delete this.
	// if len(search) == 0 {
	select {}
	// }

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
