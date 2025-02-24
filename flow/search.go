package flow

import (
	"errors"
	"io"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/server"
	"github.com/rs/zerolog/log"
)

type SearchResult struct {
	Username string
	Query    string
}

func (c *Client) GlobalSearch(token soul.Token, query string) error {
	search := new(server.FileSearch)
	searchMessage, err := search.Serialize(token, query)
	if err != nil {
		return err
	}

	_, err = c.Write(searchMessage)
	if err != nil {
		return err
	}

	log.Debug().Str("query", query).Msg("global search")

	for {
		c.mu.Lock()
		if r, ok := c.m[server.FileSearchCode]; ok {
			if len(r) > 0 {
				err = search.Deserialize(r[0])
				if err != nil && !errors.Is(err, io.EOF) {
					c.mu.Unlock()
					return err
				}

				log.Debug().Str("username", search.Username).Str("query", search.SearchQuery).Msg("search result")

				c.m[server.FileSearchCode] = c.m[server.FileSearchCode][1:]

				c.search[search.Token] = append(c.search[search.Token], SearchResult{
					Username: search.Username,
					Query:    search.SearchQuery,
				})
			}
		}
		c.mu.Unlock()

		time.Sleep(100 * time.Millisecond)
	}
}

func (c *Client) PollSearchResults(token soul.Token) []SearchResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	// if searchResult, ok := c.search[soul.Token(token)]; ok {
	// 	return searchResult
	// }
	var searchResults []SearchResult
	for _, v := range c.search {
		searchResults = append(searchResults, v...)
	}

	return searchResults
}
