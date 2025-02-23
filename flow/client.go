package flow

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/server"
	"github.com/rs/zerolog/log"
)

type Client struct {
	Config *Config

	conn net.Conn
	m    map[soul.ServerCode][]io.Reader
	mu   sync.Mutex
}

func (s *Client) Dial() (err error) {
	s.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%v", s.Config.SoulseekAddress, s.Config.SoulseekPort))
	if err != nil {
		return
	}

	s.m = make(map[soul.ServerCode][]io.Reader)

	return
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) NextMessage() (soul.ServerCode, error) {
	r, _, code, err := server.MessageRead(c.conn)
	if err != nil {
		return 0, err
	}

	c.mu.Lock()
	c.m[code] = append(c.m[code], r)
	c.mu.Unlock()

	log.Debug().Int("code", int(code)).Msg("nextmessage")

	return code, nil
}

func (c *Client) Write(message []byte) (int, error) {
	return server.MessageWrite(c.conn, message)
}

func (c *Client) Ping() error {
	ping := new(server.Ping)
	for {
		pingMessage, err := ping.Serialize()
		if err != nil {
			return err
		}

		_, err = c.Write(pingMessage)
		if err != nil {
			return err
		}

		time.Sleep(5 * time.Minute)
	}
}
