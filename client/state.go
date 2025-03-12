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

// State represents the client state.
type State struct {
	client    *Client
	searches  map[soul.Token]chan *peer.FileSearchResponse
	peers     map[string]*Peer
	mu        sync.RWMutex
	connected int64

	zerolog.Logger
}

// NewState returns a new State.
func NewState(c *Client) *State {
	s := &State{
		client:   c,
		searches: make(map[soul.Token]chan *peer.FileSearchResponse),
		peers:    make(map[string]*Peer),
	}

	s.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

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
	}()

	login := new(server.Login)

	serialized, err := login.Serialize(s.client.config.Username, s.client.config.Password)
	if err != nil {
		return err
	}

	s.client.Writer <- serialized

	s.Debug().Msg("login message sent")

	login = <-lis.Ch()

	s.Debug().Str("Greet", login.Greet).Str("IP", login.IP.String()).Msg("login message received")

	// Send the rest of login messages.
	privileges := new(server.CheckPrivileges)
	privilegesMessage, err := privileges.Serialize()
	if err != nil {
		return err
	}

	s.client.Writer <- privilegesMessage

	port := new(server.SetListenPort)
	portMessage, err := port.Serialize(uint32(s.client.config.OwnPort))
	if err != nil {
		return err
	}

	s.client.Writer <- portMessage

	status := new(server.SetStatus)
	statusMessage, err := status.Serialize(server.StatusOnline)
	if err != nil {
		return err
	}

	s.client.Writer <- statusMessage

	shared := new(server.SharedFoldersFiles)
	sharedMessage, err := shared.Serialize(s.client.config.SharedFolders, s.client.config.SharedFiles)
	if err != nil {
		return err
	}

	s.client.Writer <- sharedMessage

	watch := new(server.WatchUser)
	watchMessage, err := watch.Serialize(s.client.config.Username)
	if err != nil {
		return err
	}

	s.client.Writer <- watchMessage

	noParent := new(server.HaveNoParent)
	parentSearchMessage, err := noParent.Serialize(true)
	if err != nil {
		return err
	}

	s.client.Writer <- parentSearchMessage

	root := new(server.BranchRoot)
	rootMessage, err := root.Serialize(s.client.config.Username)
	if err != nil {
		return err
	}

	s.client.Writer <- rootMessage

	level := new(server.BranchLevel)
	levelMessage, err := level.Serialize(0)
	if err != nil {
		return err
	}

	s.client.Writer <- levelMessage

	accept := new(server.AcceptChildren)
	acceptMessage, err := accept.Serialize(true)
	if err != nil {
		return err
	}

	s.client.Writer <- acceptMessage

	s.Debug().Msg("login messages sent")

	ctxI, cancelI := context.WithTimeout(context.Background(), s.client.config.LoginTimeout)
	var i atomic.Uint32
	go func() {
		for {
			if i.Load() == 8 {
				cancelI()
				return
			}

			select {
			case r := <-room.Ch():
				i.Add(1)
				s.Debug().Int("room", len(r.Rooms)).Msg("room")

			case sp := <-speed.Ch():
				i.Add(1)
				s.Debug().Int("speed", sp.MinSpeed).Msg("speed")

			case r := <-ratio.Ch():
				i.Add(1)
				s.Debug().Int("ratio", r.SpeedRatio).Msg("ratio")

			case w := <-wish.Ch():
				i.Add(1)
				s.Debug().Int("wish", w.Interval).Msg("wish")

			case p := <-priv.Ch():
				i.Add(1)
				s.Debug().Int("priv", len(p.Users)).Msg("priv")

			case p := <-phrases.Ch():
				i.Add(1)
				s.Debug().Strs("phrases", p.Phrases).Msg("phrases")

			case o := <-ownPriv.Ch():
				i.Add(1)
				s.Debug().Int("self privilege", o.TimeLeft).Msg("ownPriv")

			case m := <-me.Ch():
				i.Add(1)

				if m.Username != s.client.config.Username {
					s.Error().Any("not me", m).Any("me", s.client.config.Username).Send()
					continue
				}

				s.Debug().Any("me", m).Send()

			case <-ctxI.Done():
				cancelI()
				return
			}
		}
	}()

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

	search := new(server.FileSearch)
	searchMessage, err := search.Serialize(token, query)
	if err != nil {
		s.Error().Err(err).Msg("search")
		return
	}

	// Send search message.
	s.client.Writer <- searchMessage

	s.Debug().Str(fmt.Sprintf("%v", token), query).Msg("search message sent")

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

type Download struct {
	Username string
	Token    soul.Token
	File     *peer.File
}

// Download sends download message to the server and listens for the responses.
func (s *State) Download(ctx context.Context, file Download) (status chan string, e chan error) {
	status = make(chan string)
	e = make(chan error)

	go func() {
		queue := new(peer.QueueUpload)
		queueMessage, err := queue.Serialize(file.File.Name)
		if err != nil {
			e <- err
			return
		}

		s.mu.RLock()
		p, ok := s.peers[file.Username]
		s.mu.RUnlock()

		if !ok {
			e <- errors.New("no peer")
			return
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					e <- errors.New("context done")
					return

				case m := <-p.initListeners.uploadDenied:
					e <- m.Reason
					return

				case <-p.initListeners.uploadFailed:
					e <- errors.New("upload failed")
					return
				}
			}
		}()

		if ok {
			sl := s.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str(fmt.Sprintf("%v", file.Token), file.File.Name).Logger()

			status <- "sending queue message"

			p.Writer <- queueMessage

			sl.Debug().Msg("queue upload")

			transfer := <-p.initListeners.transferRequest

			status <- transfer.Direction.String()

			sl.Debug().Msg("transfer request")

			response := new(peer.TransferResponse)
			responseMessage, err := response.Serialize(transfer.Token, true)
			if err != nil {
				e <- err
				return
			}

			p.Writer <- responseMessage

			status <- "response message sent"

			sl.Debug().Msg("transfer response")

			fileConn, err := p.File(ctx, transfer.Token, 0)
			if err != nil {
				e <- err
				return
			}

			status <- "file connection established"

			sl.Debug().Msg("file connection")

			var localFile *os.File
			localFile, err = os.Create(path.Join(s.client.config.DownloadFolder, file.File.Name))
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

			status <- fmt.Sprintf("file created, size: %s", humanize.Bytes(file.File.Size))

			var readSoFar int64
			for {
				n, err := io.CopyN(localFile, fileConn, 10000)
				if err != nil && !errors.Is(err, io.EOF) {
					e <- err
					return
				}

				if errors.Is(err, io.EOF) {
					break
				}

				readSoFar += n

				status <- fmt.Sprintf("copied %v%%", readSoFar*100/int64(file.File.Size))
			}

			sl.Debug().Msg("file download")

			e <- peer.ErrComplete

			return
		}
	}()

	return
}

func (s *State) max(connType soul.ConnectionType) {
	if connType == file.ConnectionType || connType == distributed.ConnectionType {
		return
	}

	for {
		s.mu.RLock()
		ok := s.connected < s.client.config.MaxPeers
		s.Debug().Int("active peer connection", int(s.connected))
		s.mu.RUnlock()

		if ok {
			break
		} else {
			time.Sleep(1 * time.Second)
			continue
		}
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

				il := s.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().
					Str("username", init.RemoteUsername).
					Str("ip", init.Conn.RemoteAddr().String()).
					Str("connection type", string(init.ConnectionType)).
					Logger()

				il.Debug().Msg("init")

				s.mu.Lock()
				p, found := s.peers[init.RemoteUsername]
				if found {
					s.mu.Unlock()

					p.Logic(init.ConnectionType, init.Conn)

					il.Debug().Msg("peer updated")
				}

				if !found {
					p = NewPeer(s.client.config, init.PeerInit, init.Conn)

					s.peers[init.RemoteUsername] = p
					s.mu.Unlock()

					il.Debug().Msg("peer added")
				}

				atomic.AddInt64(&s.connected, 1)
				go func() {
					<-p.ctx.Done()
					atomic.AddInt64(&s.connected, -1)
				}()

				// If the connection is of type P (peer), start the file response listener.
				// The previous fileResponse, if any, is cancelled in the Logic step (or NewPeer)
				// if the connection is of peer P. Thus it is safe to start a new one here.
				if init.ConnectionType == peer.ConnectionType {
					go s.fileResponse(p)
				}
			}(init)

		// Peer indirectly connects to us.
		case connect := <-connect.Ch():
			go func(connect *server.ConnectToPeer) {
				s.max(connect.Type)

				cl := s.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().
					Str("username", connect.Username).
					Str("ip", connect.IP.String()).
					Int("port", connect.Port).
					Int("obfuscated port", connect.ObfuscatedPort).
					Uint32("token", uint32(connect.Token)).
					Bool("privileged", connect.Privileged).
					Str("connection type", string(connect.Type)).
					Logger()

				cl.Debug().Msg("server connect-to-peer request")

				conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", connect.IP.String(), connect.Port))
				if err != nil {
					cl.Error().Err(err).Msg("dial")
					return
				}

				cl.Debug().Msg("connected to peer")

				firewall := new(peer.PierceFirewall)
				message, err := firewall.Serialize(connect.Token)
				if err != nil {
					cl.Error().Err(err).Msg("init")
					return
				}

				_, err = conn.Write(message)
				if err != nil {
					cl.Error().Err(err).Msg("firewall write")
					return
				}

				cl.Debug().Msg("firewall message sent")

				s.mu.Lock()
				p, ok := s.peers[connect.Username]
				if !ok {
					p = NewPeer(s.client.config, &peer.PeerInit{
						RemoteUsername: connect.Username,
						ConnectionType: connect.Type,
					}, conn)

					cl.Debug().Msg("peer added")
				}

				if ok {
					p.Logic(connect.Type, conn)

					cl.Debug().Msg("peer updated")
				}

				p.ip = connect.IP
				p.port = connect.Port

				s.peers[p.username] = p
				s.mu.Unlock()

				atomic.AddInt64(&s.connected, 1)
				go func() {
					<-p.ctx.Done()
					atomic.AddInt64(&s.connected, -1)
				}()

				if connect.Type == peer.ConnectionType {
					go s.fileResponse(p)
				}
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

				s.peers[p.username] = p
				s.mu.Unlock()
			} else {
				s.mu.Unlock()
				s.Warn().Str("status", status.Status.String()).Str("username", status.Username).Msg("peer not found")
			}

		case stats := <-statsListener.Ch():
			s.mu.Lock()
			p, ok := s.peers[stats.Username]
			if ok {
				p.averageSpeed = stats.Speed
				p.queued = stats.Uploads

				s.peers[p.username] = p
				s.mu.Unlock()
			} else {
				s.mu.Unlock()
				s.Warn().Any("stats", stats).Msg("peer not found")
			}

		case parents := <-parentsListener.Ch():
			go func(parents *server.PossibleParents) {
				pl := s.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Any("parents", parents.Parents).Logger()

				pl.Debug().Msg("init")

				// Communicate to server that it should not send us more parents.
				have := new(server.HaveNoParent)
				haveMessage, err := have.Serialize(false)
				if err != nil {
					pl.Error().Err(err).Msg("have")
					return
				}

				s.client.Writer <- haveMessage

				pl.Debug().Msg("stop receiving parents message sent")

				s.distributed(parents)

				pl.Debug().Msg("no parent connected, trying again")

				have = new(server.HaveNoParent)
				haveMessage, err = have.Serialize(true)
				if err != nil {
					s.Error().Err(err).Msg("have")
					return
				}

				s.client.Writer <- haveMessage
			}(parents)

		case watch := <-watchListener.Ch():
			s.Debug().Any("watch", watch).Msg("watch")

			s.mu.Lock()
			p, ok := s.peers[watch.Username]
			if ok {
				p.status = watch.Status
				p.averageSpeed = watch.AverageSpeed
				p.queued = watch.UploadNumber

				s.peers[p.username] = p
				s.mu.Unlock()
			} else {
				s.mu.Unlock()
				s.Warn().Any("watch", watch).Msg("peer not found")
			}
		}
	}
}

func (s *State) distributed(m *server.PossibleParents) {
	for _, parent := range m.Parents {
		pl := s.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str("parent", parent.Username).Logger()

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
				RemoteUsername: parent.Username,
				ConnectionType: distributed.ConnectionType,
			}, conn)

			s.peers[p.username] = p
			s.mu.Unlock()

			pl.Debug().Msg("peer added")
		}

		if ok {
			s.mu.Unlock()

			p.Logic(distributed.ConnectionType, conn)

			pl.Debug().Msg("peer updated")
		}

		init := new(peer.PeerInit)
		message, err := init.Serialize(s.client.config.Username, distributed.ConnectionType)
		if err != nil {
			pl.Error().Err(err).Msg("init")
			continue
		}

		p.distributedWriter <- message

		pl.Info().Msg("parent connected")

		for {
			pl.Debug().Msg("waiting for parent")
			select {
			case branch := <-p.initDistributedListeners.branchRoot:
				pl.Debug().Any("branch", branch).Msg("branch")

			case level := <-p.initDistributedListeners.branchLevel:
				pl.Debug().Any("level", level).Msg("level")

			case embed := <-p.initDistributedListeners.embeddedMessage:
				pl.Debug().Any("embed", embed).Msg("embed")

			case search := <-p.initDistributedListeners.search:
				pl.Debug().Any("search", search).Msg("search")
			}
		}

		break
	}
}

// fileResponse listens for file search responses from a peer and passes them on to the internal
// searches channel.
func (s *State) fileResponse(p *Peer) {
	for {
		select {
		case <-p.ctx.Done():
			return

		case fileResponse := <-p.initListeners.fileSearchResponse:
			s.mu.RLock()
			channel, ok := s.searches[fileResponse.Token]
			s.mu.RUnlock()

			switch ok {
			case true:
				channel <- fileResponse

			case false:
				s.Debug().Any("message", fileResponse).Msg("search not found")
			}
		}
	}
}
