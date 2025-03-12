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
	"github.com/rs/zerolog/log"
)

// State represents the client state.
type State struct {
	client    *Client
	searches  map[soul.Token]chan *peer.FileSearchResponse
	peers     map[string]*Peer
	mu        sync.RWMutex
	connected int64
}

// NewState returns a new State.
func NewState(c *Client) *State {
	s := &State{
		client:   c,
		searches: make(map[soul.Token]chan *peer.FileSearchResponse),
		peers:    make(map[string]*Peer),
	}

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

	log.Debug().Msg("login message sent")

	login = <-lis.Ch()

	log.Debug().Str("Greet", login.Greet).Str("IP", login.IP.String()).Msg("login message received")

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

	log.Debug().Msg("login messages sent")

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
				log.Debug().Int("room", len(r.Rooms)).Msg("room")

			case s := <-speed.Ch():
				i.Add(1)
				log.Debug().Any("speed", s).Msg("speed")

			case r := <-ratio.Ch():
				i.Add(1)
				log.Debug().Any("ratio", r).Msg("ratio")

			case w := <-wish.Ch():
				i.Add(1)
				log.Debug().Any("wish", w).Msg("wish")

			case p := <-priv.Ch():
				i.Add(1)
				log.Debug().Int("priv", len(p.Users)).Msg("priv")

			case p := <-phrases.Ch():
				i.Add(1)
				log.Debug().Any("phrases", p).Msg("phrases")

			case o := <-ownPriv.Ch():
				i.Add(1)
				log.Debug().Any("ownPriv", o).Msg("ownPriv")

			case m := <-me.Ch():
				i.Add(1)

				if m.Username != s.client.config.Username {
					log.Error().Any("me", m).Msg("not me")
					continue
				}

				log.Debug().Any("me", m).Msg("me")

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
		log.Error().Err(err).Msg("search")
		return
	}

	// Send search message.
	s.client.Writer <- searchMessage

	log.Debug().Msg("search message sent")

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

		// go func() {
		// 	for {
		// 		select {
		// 		case <-ctx.Done():
		// 			e <- errors.New("context done")
		// 			return
		// 		}

		// 	}
		// 	log.Fatal().Any("m", *<-p.initListeners.uploadFailed).Send()
		// }()

		if ok {
			status <- "sending queue message"

			p.Writer <- queueMessage

			transfer := <-p.initListeners.transferRequest

			status <- transfer.Direction.String()

			log.Debug().Str("username", file.Username).Uint32("token", uint32(file.Token)).Str("file", file.File.Name).Msg("transfer request")

			response := new(peer.TransferResponse)
			responseMessage, err := response.Serialize(transfer.Token, true)
			if err != nil {
				e <- err
				return
			}

			p.Writer <- responseMessage

			status <- "response message sent"

			log.Debug().Str("username", file.Username).Uint32("token", uint32(file.Token)).Str("file", file.File.Name).Msg("transfer response")

			fileConn, err := p.File(ctx, transfer.Token, 0)
			if err != nil {
				e <- err
				return
			}

			status <- "file connection established"

			log.Debug().Str("username", file.Username).Uint32("token", uint32(file.Token)).Str("file", file.File.Name).Msg("file connection")

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

			status <- fmt.Sprintf("file created %v", file.File.Size)

			n, err := io.CopyN(localFile, fileConn, int64(file.File.Size))
			if err != nil && !errors.Is(err, io.EOF) {
				e <- err
				return
			}

			status <- fmt.Sprintf("copied %v", n)

			log.Debug().Str("username", file.Username).Uint32("token", uint32(file.Token)).Str("file", file.File.Name).Msg("file download")

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
		log.Debug().Int("active peer connection", int(s.connected))
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
			go func(message *PeerInit) {
				s.max(init.ConnectionType)

				log.Debug().Any("init", init).Send()

				s.mu.RLock()
				p, found := s.peers[message.RemoteUsername]
				s.mu.RUnlock()

				if found {
					p.Logic(message.ConnectionType, message.Conn)

					log.Debug().Str("username", message.RemoteUsername).Msg("peer updated")
				}

				if !found {
					p = NewPeer(s.client.config, message.PeerInit, message.Conn)

					s.mu.Lock()
					s.peers[message.RemoteUsername] = p
					s.mu.Unlock()

					log.Debug().Str("username", message.RemoteUsername).Msg("peer added")
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

				log.Debug().Any("connect", connect).Send()

				conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", connect.IP.String(), connect.Port))
				if err != nil {
					atomic.AddInt64(&s.connected, -1)
					return
				}

				log.Debug().Any("connect", connect).Msg("connected")

				s.mu.RLock()
				p, ok := s.peers[connect.Username]
				s.mu.RUnlock()

				if !ok {
					p = NewPeer(s.client.config, &peer.PeerInit{
						RemoteUsername: connect.Username,
						ConnectionType: peer.ConnectionType,
					}, conn)

					p.ip = connect.IP
					p.port = connect.Port

					s.mu.Lock()
					s.peers[p.username] = p
					s.mu.Unlock()

					log.Debug().Str("username", connect.Username).Msg("peer added")
				}

				if ok {
					p.Logic(distributed.ConnectionType, conn)

					s.mu.Lock()
					p.ip = connect.IP
					p.port = connect.Port
					s.peers[p.username] = p
					s.mu.Unlock()

					log.Debug().Str("username", connect.Username).Msg("peer updated")
				}

				atomic.AddInt64(&s.connected, 1)
				go func() {
					<-p.ctx.Done()
					atomic.AddInt64(&s.connected, -1)
				}()

				go s.fileResponse(p)

				firewall := new(peer.PierceFirewall)
				message, err := firewall.Serialize(connect.Token)
				if err != nil {
					log.Error().Err(err).Msg("init")
					return
				}

				p.Writer <- message

				log.Debug().Any("connect", connect).Msg("firewall message sent")

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
			s.mu.RLock()
			p, ok := s.peers[status.Username]
			s.mu.RUnlock()

			if ok {
				p.mu.Lock()
				p.status = status.Status
				p.privileged = status.Privileged
				p.mu.RUnlock()
			}

		case stats := <-statsListener.Ch():
			s.mu.RLock()
			p, ok := s.peers[stats.Username]
			s.mu.RUnlock()

			if ok {
				p.mu.Lock()
				p.averageSpeed = stats.Speed
				p.queued = stats.Uploads
				p.mu.RUnlock()
			}

		case parents := <-parentsListener.Ch():
			go func(parents *server.PossibleParents) {
				log.Debug().Any("parents", parents.Parents).Msg("init")

				// Communicate to server that it should not send us more parents.
				have := new(server.HaveNoParent)
				haveMessage, err := have.Serialize(false)
				if err != nil {
					log.Error().Err(err).Msg("have")
					return
				}

				s.client.Writer <- haveMessage

				log.Debug().Msg("stop receiving parents message sent")

				s.distributed(parents)

				log.Debug().Any("message", parents).Msg("no parent connected, trying again")

				have = new(server.HaveNoParent)
				haveMessage, err = have.Serialize(true)
				if err != nil {
					log.Error().Err(err).Msg("have")
					return
				}

				s.client.Writer <- haveMessage
			}(parents)

		case watch := <-watchListener.Ch():
			log.Debug().Any("watch", watch).Msg("watch")

			s.mu.RLock()
			p, ok := s.peers[watch.Username]
			s.mu.RUnlock()

			if ok {
				p.mu.Lock()
				p.status = watch.Status
				p.averageSpeed = watch.AverageSpeed
				p.queued = watch.UploadNumber
				p.mu.RUnlock()
			}

			if !ok {
				log.Warn().Any("watch", watch).Msg("peer not found")
			}
		}
	}
}

func (s *State) distributed(m *server.PossibleParents) {
	for _, parent := range m.Parents {
		log.Debug().Any("parent", parent).Msg("trying parent")

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", parent.IP.String(), parent.Port))
		if err != nil {
			log.Error().Any("message", m).Err(err).Msg("distributed")
			continue
		}

		log.Debug().Any("parent", parent).Msg("connected")

		s.mu.RLock()
		p, ok := s.peers[parent.Username]
		s.mu.RUnlock()
		if !ok {
			p = NewPeer(s.client.config, &peer.PeerInit{
				RemoteUsername: parent.Username,
				ConnectionType: distributed.ConnectionType,
			}, conn)

			s.mu.Lock()
			s.peers[p.username] = p
			s.mu.Unlock()

			log.Debug().Str("username", parent.Username).Msg("peer added")
		}

		if ok {
			p.Logic(distributed.ConnectionType, conn)

			log.Debug().Str("username", parent.Username).Msg("peer updated")
		}

		init := new(peer.PeerInit)
		message, err := init.Serialize(s.client.config.Username, distributed.ConnectionType)
		if err != nil {
			log.Error().Any("message", m).Err(err).Msg("init")
			continue
		}

		p.distributedWriter <- message

		log.Info().Any("parent", parent).Msg("parent connected")

		for {
			log.Debug().Any("parent", parent).Msg("waiting for parent")
			select {
			case branch := <-p.initDistributedListeners.branchRoot:
				log.Debug().Any("branch", branch).Msg("branch")

			case level := <-p.initDistributedListeners.branchLevel:
				log.Debug().Any("level", level).Msg("level")

			case embed := <-p.initDistributedListeners.embeddedMessage:
				log.Debug().Any("embed", embed).Msg("embed")

			case search := <-p.initDistributedListeners.search:
				log.Debug().Any("search", search).Msg("search")
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
				log.Debug().Any("message", fileResponse).Msg("search not found")
			}
		}
	}
}
