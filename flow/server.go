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

type Server struct {
	Config *Config

	conn net.Conn
	m    map[soul.ServerCode][]io.Reader
	mu   sync.Mutex
}

func (s *Server) Dial() (err error) {
	s.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%v", s.Config.SoulseekAddress, s.Config.SoulseekPort))
	if err != nil {
		return
	}

	s.m = make(map[soul.ServerCode][]io.Reader)

	return
}

func (s *Server) Close() {
	s.conn.Close()
}

func (s *Server) NextMessage() (soul.ServerCode, error) {
	r, _, code, err := server.MessageRead(s.conn)
	if err != nil {
		return 0, err
	}

	s.mu.Lock()
	s.m[code] = append(s.m[code], r)
	s.mu.Unlock()

	log.Debug().Int("code", int(code)).Msg("nextmessage")

	return code, nil
}

func (s *Server) Write(message []byte) (int, error) {
	return server.MessageWrite(s.conn, message)
}

func (s *Server) Ping() error {
	ping := new(server.Ping)
	for {
		pingMessage, err := ping.Serialize()
		if err != nil {
			return err
		}

		_, err = s.Write(pingMessage)
		if err != nil {
			return err
		}

		time.Sleep(5 * time.Minute)
	}
}
