package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/distributed"
	"github.com/bh90210/soul/file"
	"github.com/bh90210/soul/peer"
	"github.com/bh90210/soul/server"
	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// HundredKb 100Kb is the size of the buffer for file downloads.
const HundredKb = 100000

// State represents the client state.
type State struct {
	client    *Client
	searches  map[soul.Token]chan *peer.FileSearchResponse
	peers     map[string]*Peer
	mu        sync.RWMutex
	connected int64
	log       zerolog.Logger
}

// NewState returns a new State.
func NewState(c *Client) *State {
	s := &State{
		client:   c,
		searches: make(map[soul.Token]chan *peer.FileSearchResponse),
		peers:    make(map[string]*Peer),
	}

	s.log = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	return s
}

// Login sends login message to the server and listens for the responses.
func (s *State) Login(ctx context.Context) error {
	lis := s.client.Relays.Login.Listener(0)
	room := s.client.Relays.RoomList.Listener(0)
	speed := s.client.Relays.ParentMinSpeed.Listener(0)
	ratio := s.client.Relays.ParentSpeedRatio.Listener(0)
	wish := s.client.Relays.WishlistInterval.Listener(0)
	priv := s.client.Relays.PrivilegedUsers.Listener(0)
	phrases := s.client.Relays.ExcludedSearchPhrases.Listener(0)
	ownPriv := s.client.Relays.CheckPrivileges.Listener(0)
	me := s.client.Relays.WatchUser.Listener(0)
	ownAddress := s.client.Relays.GetPeerAddress.Listener(0)

	defer func() {
		lis.Close()
		room.Close()
		speed.Close()
		ratio.Close()
		wish.Close()
		priv.Close()
		phrases.Close()
		ownPriv.Close()
		me.Close()
		ownAddress.Close()
	}()

	ctxI, cancelI := context.WithTimeout(context.Background(), s.client.config.LoginTimeout)
	var i atomic.Uint32
	go func() {
		for {
			if i.Load() == 9 {
				go cancelI()
				return
			}

			select {
			case r := <-room.Ch():
				i.Add(1)
				s.log.Debug().Int("room", len(r.Rooms)).Uint32("i", i.Load()).Msg("room")

			case sp := <-speed.Ch():
				i.Add(1)
				s.log.Debug().Int("speed", sp.MinSpeed).Uint32("i", i.Load()).Msg("speed")

			case r := <-ratio.Ch():
				i.Add(1)
				s.log.Debug().Int("ratio", r.SpeedRatio).Uint32("i", i.Load()).Msg("ratio")

			case w := <-wish.Ch():
				i.Add(1)
				s.log.Debug().Int("wish", w.Interval).Uint32("i", i.Load()).Msg("wish")

			case p := <-priv.Ch():
				i.Add(1)
				s.log.Debug().Int("priv", len(p.Users)).Uint32("i", i.Load()).Msg("priv")

			case p := <-phrases.Ch():
				i.Add(1)
				s.log.Debug().Strs("phrases", p.Phrases).Uint32("i", i.Load()).Msg("phrases")

			case o := <-ownPriv.Ch():
				i.Add(1)
				s.log.Debug().Int("self privilege", o.TimeLeft).Uint32("i", i.Load()).Msg("ownPriv")

			case m := <-me.Ch():
				i.Add(1)

				if m.Username != s.client.config.Username {
					s.log.Error().Any("not me", m).Uint32("i", i.Load()).Any("me", s.client.config.Username).Send()
					continue
				}

				s.log.Debug().Any("me", m).Uint32("i", i.Load()).Send()

			case o := <-ownAddress.Ch():
				i.Add(1)
				s.log.Debug().Any("own address", o).Uint32("i", i.Load()).Msg("own address")

			case <-ctxI.Done():
				go cancelI()
				return
			}
		}
	}()

	_, err := server.Write(s.client.Conn(), &server.Login{Username: s.client.config.Username, Password: s.client.config.Password})
	if err != nil {
		return err
	}

	s.log.Debug().Msg("login message sent")

	login := <-lis.Ch()

	s.log.Debug().Str("Greet", login.Greet).Str("IP", login.IP.String()).Msg("login message received")

	// Send the rest of login messages.
	_, err = server.Write(s.client.Conn(), &server.CheckPrivileges{})
	if err != nil {
		return err
	}

	port := &server.SetListenPort{Port: s.client.config.OwnPort}
	if s.client.config.OwnObfuscatedPort != 0 {
		port.ObfuscatedPort = s.client.config.OwnObfuscatedPort
	}

	_, err = server.Write(s.client.Conn(), port)
	if err != nil {
		return err
	}

	_, err = server.Write(s.client.Conn(), &server.SetStatus{Status: server.StatusOnline})
	if err != nil {
		return err
	}

	_, err = server.Write(s.client.Conn(), &server.SharedFoldersFiles{Directories: s.client.config.SharedFolders, Files: s.client.config.SharedFiles})
	if err != nil {
		return err
	}

	_, err = server.Write(s.client.Conn(), &server.WatchUser{Username: s.client.config.Username})
	if err != nil {
		return err
	}

	_, err = server.Write(s.client.Conn(), &server.HaveNoParent{Have: true})
	if err != nil {
		return err
	}

	_, err = server.Write(s.client.Conn(), &server.BranchRoot{Root: s.client.config.Username})
	if err != nil {
		return err
	}

	_, err = server.Write(s.client.Conn(), &server.BranchLevel{Level: 0})
	if err != nil {
		return err
	}

	_, err = server.Write(s.client.Conn(), &server.AcceptChildren{Accept: true})
	if err != nil {
		return err
	}

	_, err = server.Write(s.client.Conn(), &server.GetPeerAddress{Username: s.client.config.Username})
	if err != nil {
		return err
	}

	s.log.Debug().Msg("login messages sent")

	<-ctxI.Done()

	go s.peer(ctx)
	go s.server(ctx)

	return nil
}

// Search sends search message to the server and listens for the responses.
func (s *State) Search(ctx context.Context, query string, token soul.Token) (results chan *peer.FileSearchResponse, err error) {
	results = make(chan *peer.FileSearchResponse)

	s.mu.Lock()
	s.searches[token] = results
	s.mu.Unlock()

	s.client.mu.RLock()
	_, err = server.Write(s.client.Conn(), &server.FileSearch{Token: token, SearchQuery: query})
	s.client.mu.RUnlock()
	if err != nil {
		s.log.Error().Err(err).Msg("search")
		return
	}

	s.log.Debug().Str(fmt.Sprintf("%v", token), query).Msg("search message sent")

	go func() {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			delete(s.searches, token)
			s.mu.Unlock()
			return
		}
	}()

	return
}

// Download sends download message to the server and listens for the responses.
func (s *State) Download(ctx context.Context, files *peer.FileSearchResponse) (status chan string, e chan error) {
	status = make(chan string)
	e = make(chan error)

	go func() {
		s.mu.RLock()
		p, ok := s.peers[files.Username]
		s.mu.RUnlock()

		if !ok {
			e <- errors.New("no peer")
			return
		}

		tRequest := p.Relays.TransferRequest.Listener(1)
		defer tRequest.Close()

		failed := p.Relays.UploadFailed.Listener(1)
		defer failed.Close()

		denied := p.Relays.UploadDenied.Listener(1)
		defer denied.Close()

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					e <- errors.New("context done")
					return

				case <-failed.Ch():
					e <- errors.New("upload failed")
					continue

				case m := <-denied.Ch():
					e <- m.Reason
					return
				}
			}
		}()

		if ok {
			sl := s.log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str(fmt.Sprintf("%v", files.Token), files.Results[0].Name).Logger()

			sl.Debug().Msg("queue upload")

			conn, obfuscated := p.Conn(peer.ConnectionType)

			_, err := peer.Write(conn, &peer.QueueUpload{Filename: files.Results[0].Name}, obfuscated)
			if err != nil {
				e <- err
				return
			}

			status <- "queued"

			sl.Debug().Msg("waiting transfer request")

			transfer := <-tRequest.Ch()

			status <- "starting"

			sl.Debug().Msg("transfer response")

			_, err = peer.Write(conn, &peer.TransferResponse{Token: transfer.Token, Allowed: true}, obfuscated)
			if err != nil {
				e <- err
				return
			}

			sl.Debug().Msg("creating local file")

			var localFile *os.File
			localFile, err = os.Create(path.Join(s.client.config.DownloadFolder, files.Results[0].Name))
			if err != nil {
				e <- err
				return
			}

			defer localFile.Close()

			_, err = localFile.Seek(int64(0), 0)
			if err != nil {
				e <- err
				return
			}

			status <- fmt.Sprintf("file created, size: %s", humanize.Bytes(files.Results[0].Size))

			status <- "waiting for F connection"

			var fileConn net.Conn
			for {
				select {
				case <-ctx.Done():
					e <- errors.New("context done")
					return

				default:
					fileConn, _ = p.Conn(file.ConnectionType, transfer.Token)
					if fileConn == nil {
						time.Sleep(time.Second)
					}
				}

				if fileConn != nil {
					break
				}
			}

			status <- "sending offset"

			_, err = file.Write(fileConn, &file.Offset{Offset: 0})
			if err != nil {
				e <- err
				return
			}

			var readSoFar int64
			for {
				n, err := io.CopyN(localFile, fileConn, HundredKb)
				if err != nil && !errors.Is(err, io.EOF) {
					e <- err
					return
				}

				if errors.Is(err, io.EOF) {
					status <- fmt.Sprint("copied 100%")
					break
				}

				readSoFar += n

				status <- fmt.Sprintf("copied %v%%", readSoFar*100/int64(files.Results[0].Size))
			}

			sl.Debug().Msg("file download")

			// e <- peer.ErrComplete

			wg.Wait()

			return
		}
	}()

	return
}

func (s *State) max(connType soul.ConnectionType) {
	switch connType {
	case peer.ConnectionType:
		for {
			s.mu.RLock()
			ok := s.connected < s.client.config.MaxPeers
			s.mu.RUnlock()

			if ok {
				break
			} else {
				time.Sleep(1 * time.Second)
				continue
			}
		}

	case file.ConnectionType, distributed.ConnectionType:
		return
	}
}

// peer covers the three ways peers can start a connection with us.
func (s *State) peer(ctx context.Context) {
	connect := s.client.Relays.ConnectToPeer.Listener(1)
	defer connect.Close()

	for {
		select {
		case <-ctx.Done():
			return

		// We made an indirect connection request to a peer.
		// case firewall := <-s.client.Firewall:

		// Peer directly connects to us.
		case init := <-s.client.Init:
			go func(init *PeerInit) {
				s.max(init.ConnectionType)

				il := s.log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().
					Str("username", init.Username).
					Str("ip", init.Conn.RemoteAddr().String()).
					Str("connection type", string(init.ConnectionType)).
					Bool("obfuscated", init.Obfuscated).
					Logger()

				il.Debug().Msg("init")

				s.mu.Lock()
				p, found := s.peers[init.Username]
				if !found {
					p = NewPeer(s.client.config, init.PeerInit)
					s.peers[init.Username] = p
					il.Debug().Msg("peer added")
				}
				s.mu.Unlock()

				// If the connection is of type P (peer), start the file response listener.
				// The previous fileResponse, if any, is cancelled in the Logic step (or NewPeer)
				// if the connection is of peer P. Thus it is safe to start a new one here.
				if init.ConnectionType == peer.ConnectionType {
					go s.fileResponse(p)
				}

				p.New(init.ConnectionType, init.Conn, init.Obfuscated)

				il.Debug().Msg("peer updated")

				atomic.AddInt64(&s.connected, 1)
				go func() {
					<-p.ctx.Done()
					atomic.AddInt64(&s.connected, -1)
				}()
			}(init)

		// Peer indirectly connects to us.
		case connect := <-connect.Ch():
			go func(connect *server.ConnectToPeer) {
				s.max(connect.Type)

				var useObfuscatedPort bool
				if connect.Type == peer.ConnectionType {
					useObfuscatedPort = connect.ObfuscatedPort != 0
				}

				cl := s.log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().
					Str("username", connect.Username).
					Str("ip", connect.IP.String()).
					Int("port", connect.Port).
					Int("obfuscated port", connect.ObfuscatedPort).
					Bool("obfuscated", useObfuscatedPort).
					Uint32("token", uint32(connect.Token)).
					Bool("privileged", connect.Privileged).
					Str("connection type", string(connect.Type)).
					Logger()

				cl.Debug().Msg("server connect-to-peer request")

				var port int
				if useObfuscatedPort {
					port = connect.ObfuscatedPort
				} else {
					port = connect.Port
				}

				conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", connect.IP.String(), port))
				if err != nil {
					cl.Error().Err(err).Msg("dial")
					return
				}

				cl.Debug().Msg("connected to peer")

				_, err = peer.Write(conn, &peer.PierceFirewall{Token: connect.Token}, useObfuscatedPort)
				if err != nil {
					cl.Error().Err(err).Msg("pierce firewall")
					return
				}

				cl.Debug().Msg("firewall message sent")

				s.mu.Lock()
				p, ok := s.peers[connect.Username]
				if !ok {
					p = NewPeer(s.client.config, &peer.PeerInit{
						Username:       connect.Username,
						ConnectionType: connect.Type,
					})

					cl.Debug().Msg("peer added")
				}

				p.ip = connect.IP
				p.port = connect.Port
				p.obfuscatedPort = connect.ObfuscatedPort

				s.peers[p.username] = p
				s.mu.Unlock()

				if connect.Type == peer.ConnectionType {
					go s.fileResponse(p)
				}

				p.New(connect.Type, conn, useObfuscatedPort)

				cl.Debug().Msg("peer updated")

				atomic.AddInt64(&s.connected, 1)
				go func() {
					<-p.ctx.Done()
					atomic.AddInt64(&s.connected, -1)
				}()
			}(connect)
		}
	}
}

func (s *State) server(ctx context.Context) {
	statusListener := s.client.Relays.GetUserStatus.Listener(1)
	defer statusListener.Close()

	statsListener := s.client.Relays.GetUserStats.Listener(1)
	defer statsListener.Close()

	parentsListener := s.client.Relays.PossibleParents.Listener(1)
	defer parentsListener.Close()

	watchListener := s.client.Relays.WatchUser.Listener(1)
	defer watchListener.Close()

	connect := s.client.Relays.ConnectToPeer.Listener(1)
	defer connect.Close()

	for {
		select {
		case <-ctx.Done():
			return

		case status := <-statusListener.Ch():
			s.mu.Lock()
			p, ok := s.peers[status.Username]
			if ok {
				p.status = status.Status
				p.privileged = status.Privileged
				s.mu.Unlock()
			} else {
				s.mu.Unlock()
				s.log.Warn().Str("status", status.Status.String()).Str("username", status.Username).Msg("peer not found")
			}

		case stats := <-statsListener.Ch():
			s.mu.Lock()
			p, ok := s.peers[stats.Username]
			if ok {
				p.averageSpeed = stats.Speed
				p.queued = stats.Uploads
				s.mu.Unlock()
			} else {
				s.mu.Unlock()
				s.log.Warn().Any("stats", stats).Msg("peer not found")
			}

		case parents := <-parentsListener.Ch():
			go func(parents *server.PossibleParents) {
				pl := s.log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Any("parents", parents.Parents).Logger()

				pl.Debug().Msg("init")

				// Communicate to server that it should not send us more parents.
				s.mu.RLock()
				_, err := server.Write(s.client.Conn(), &server.HaveNoParent{Have: false})
				s.mu.RUnlock()
				if err != nil {
					pl.Error().Err(err).Msg("have")
					return
				}

				pl.Debug().Msg("stop receiving parents from server")

				s.distributed(parents)

				pl.Debug().Msg("no parent connected, trying again")

				s.mu.RLock()
				_, err = server.Write(s.client.Conn(), &server.HaveNoParent{Have: true})
				s.mu.RUnlock()
				if err != nil {
					pl.Error().Err(err).Msg("have")
					return
				}
			}(parents)

		case watch := <-watchListener.Ch():
			s.log.Debug().Any("watch", watch).Msg("watch")

			s.mu.Lock()
			p, ok := s.peers[watch.Username]
			if ok {
				p.status = watch.Status
				p.averageSpeed = watch.AverageSpeed
				p.queued = watch.UploadNumber
				s.mu.Unlock()
			} else {
				s.mu.Unlock()
				s.log.Warn().Any("watch", watch).Msg("peer not found")
			}
		}
	}
}

// TODO: finish me Byron.
func (s *State) distributed(m *server.PossibleParents) {
	for _, parent := range m.Parents {
		pl := s.log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str("parent", parent.Username).Logger()

		pl.Debug().Msg("trying parent")

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", parent.IP.String(), parent.Port))
		if err != nil {
			pl.Error().Err(err).Msg("distributed")
			continue
		}

		pl.Debug().Msg("connected")

		s.mu.Lock()
		p, ok := s.peers[parent.Username]
		if !ok {
			p = NewPeer(s.client.config, &peer.PeerInit{
				Username:       parent.Username,
				ConnectionType: distributed.ConnectionType,
			})

			s.peers[p.username] = p

			pl.Debug().Msg("peer added")
		}
		s.mu.Unlock()

		branch := p.Relays.Distributed.BranchRoot.Listener(1)
		defer branch.Close()

		level := p.Relays.Distributed.BranchLevel.Listener(1)
		defer level.Close()

		embed := p.Relays.Distributed.EmbeddedMessage.Listener(1)
		defer embed.Close()

		search := p.Relays.Distributed.Search.Listener(1)
		defer search.Close()

		p.New(distributed.ConnectionType, conn, false)

		pl.Debug().Msg("peer updated")

		connD, _ := p.Conn(distributed.ConnectionType)
		_, err = distributed.Write(connD, &peer.PeerInit{
			Username:       s.client.config.Username,
			ConnectionType: distributed.ConnectionType,
		})
		if err != nil {
			pl.Error().Err(err).Msg("init")
			continue
		}

		pl.Info().Msg("parent connected")

		for {
			pl.Debug().Msg("waiting for parent")
			select {
			case branch := <-branch.Ch():
				pl.Debug().Any("branch", branch).Msg("branch")

			case level := <-level.Ch():
				pl.Debug().Any("level", level).Msg("level")

			case embed := <-embed.Ch():
				pl.Debug().Any("embed", embed).Msg("embed")

			case search := <-search.Ch():
				pl.Debug().Any("search", search).Msg("search")
			}
		}
	}
}

// fileResponse listens for file search responses from a peer and passes them on to the internal
// searches channel.
func (s *State) fileResponse(p *Peer) {
	response := p.Relays.FileSearchResponse.Listener(1)
	defer response.Close()

	for {
		if p.ctx == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}

	for {
		select {
		case <-p.ctx.Done():
			return

		case fileResponse := <-response.Ch():
			s.mu.RLock()
			channel, ok := s.searches[fileResponse.Token]
			s.mu.RUnlock()

			switch ok {
			case true:
				channel <- fileResponse

			case false:
				s.log.Debug().Any("message", fileResponse).Msg("search not found")
			}
		}
	}
}
