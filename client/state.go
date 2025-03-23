package client

import (
	"bytes"
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

	"slices"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/distributed"
	"github.com/bh90210/soul/file"
	"github.com/bh90210/soul/peer"
	"github.com/bh90210/soul/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// HundredKb 100Kb is the size of the buffer for file downloads.
const HundredKb = 100000

// State represents the client state.
type State struct {
	client   *Client
	searches map[soul.Token]chan *File
	peers    map[string]*Peer // TODO: Periodically empty.
	mu       sync.RWMutex

	connectedP int64
	connectedF int64

	level    int32
	root     string
	parent   *Peer
	children []*Peer

	log zerolog.Logger
}

// NewState returns a new State.
func NewState(c *Client) *State {
	s := &State{
		client:   c,
		searches: make(map[soul.Token]chan *File),
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
					s.log.Warn().Any("not me", m).Uint32("i", i.Load()).Any("me", s.client.config.Username).Send()
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

	s.log.Info().Str("Greet", login.Greet).Str("IP", login.IP.String()).Msg("login message received")

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

type File struct {
	Username string
	Token    soul.Token
	Queue    int
	*peer.File
}

// Search sends search message to the server and listens for the responses.
func (s *State) Search(ctx context.Context, query string, token soul.Token) (results chan *File, err error) {
	results = make(chan *File)

	s.mu.Lock()
	s.searches[token] = results
	s.mu.Unlock()

	s.client.mu.RLock()
	_, err = server.Write(s.client.Conn(), &server.FileSearch{Token: token, SearchQuery: query})
	s.client.mu.RUnlock()
	if err != nil {
		s.log.Warn().Err(err).Msg("search")
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

type Status string

const (
	StatusQueued   Status = "queued"
	StatusStarting Status = "starting"
	StatusReceived Status = "received"
)

// ErrNoPeer is returned when the peer is not found.
var ErrNoPeer = errors.New("no peer")

// Download sends download message to the server and listens for the responses.
func (s *State) Download(ctx context.Context, f *File) (status chan string, e chan error) {
	// Set the status and error channels.
	status = make(chan string, 10)
	e = make(chan error, 1)

	// Try find the the username of the file to download among the peers.
	s.mu.RLock()
	p, found := s.peers[f.Username]
	s.mu.RUnlock()

	// If the peer is not found, return an error.
	if !found {
		e <- ErrNoPeer
		return
	}

	// Init peer listeners relating to the file transfer.
	tRequest := p.Relays.TransferRequest.Listener(1)
	defer tRequest.Close()

	failed := p.Relays.UploadFailed.Listener(1)
	defer failed.Close()

	denied := p.Relays.UploadDenied.Listener(1)
	defer denied.Close()

	go func() {
		for {
			select {
			case <-ctx.Done():
				e <- errors.New("context done")
				return

			case <-failed.Ch():
				info, err := os.Stat(path.Join(s.client.config.DownloadFolder, f.Name))
				if err != nil {
					if !os.IsNotExist(err) {
						e <- err
						return
					}
				}

				if info != nil {
					if info.Size() == int64(f.Size) {
						e <- peer.ErrComplete
						return
					}
				}

				e <- errors.New("failed")
				return

			case m := <-denied.Ch():
				e <- m.Reason
				return
			}
		}
	}()

	if found {
		sl := s.log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str(fmt.Sprintf("%v", f.Token), f.Name).Logger()

		sl.Debug().Msg("queue upload")

		// Send the peer the queue upload message.
		conn, obfuscated := p.Conn(peer.ConnectionType)
		_, err := peer.Write(conn, &peer.QueueUpload{Filename: f.Name}, obfuscated)
		if err != nil {
			e <- err
			return
		}

		status <- string(StatusQueued)

		sl.Debug().Msg("waiting transfer request")

		// When peer is ready to start the file transfer, it sends a transfer request.
		transfer := <-tRequest.Ch()

		sl.Debug().Msg("transfer response")

		// We reply to the transfer request with a transfer response.
		_, err = peer.Write(conn, &peer.TransferResponse{Token: transfer.Token, Allowed: true}, obfuscated)
		if err != nil {
			e <- err
			return
		}

		// Stat for the destination file.
		info, err := os.Stat(path.Join(s.client.config.DownloadFolder, f.Name))
		if err != nil {
			if !os.IsNotExist(err) {
				e <- err
				return
			}
		}

		var localFile *os.File
		// If file does not exist, create it and pass 0 to the offset.
		if os.IsNotExist(err) {
			localFile, err = os.Create(path.Join(s.client.config.DownloadFolder, f.Name))
			if err != nil {
				e <- err
				return
			}

			defer localFile.Close()

		} else {
			// If file exists count the length and pass it to the offset.
			localFile, err = os.OpenFile(path.Join(s.client.config.DownloadFolder, f.Name), os.O_RDWR, 0644)
			if err != nil {
				e <- err
				return
			}

			info, err = localFile.Stat()
			if err != nil {
				e <- err
				return
			}

			_, err = localFile.Seek(0, io.SeekEnd)
			if err != nil {
				e <- err
				return
			}
		}

		status <- string(StatusStarting)

		var fileConn net.Conn
		for {
			select {
			case <-ctx.Done():
				e <- errors.New("context done before file F connection")
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

		s.log.Debug().Int64("offset", info.Size()).Msg("sending offset")

		_, err = file.Write(fileConn, &file.Offset{Offset: uint64(info.Size())})
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
				break
			}

			readSoFar += n

			status <- fmt.Sprintf("%v%%", readSoFar*100/int64(f.Size))
		}

		sl.Debug().Msg("file download")

		status <- string(StatusReceived)

		return
	}

	return
}

func (s *State) count(ctx context.Context, connType soul.ConnectionType, p *Peer) {
	switch connType {
	case peer.ConnectionType:
		atomic.AddInt64(&s.connectedP, 1)

		go func() {
			<-ctx.Done()
			atomic.AddInt64(&s.connectedP, -1)
		}()

	case distributed.ConnectionType:
		s.mu.Lock()
		s.children = append(s.children, p)
		s.mu.Unlock()

		go func() {
			<-ctx.Done()
			s.mu.Lock()
			for k, v := range s.children {
				if v.username == p.username {
					s.children = slices.Delete(s.children, k, k+1)
					break
				}
			}
			s.mu.Unlock()
		}()

	case file.ConnectionType:
		atomic.AddInt64(&s.connectedF, 1)

		go func() {
			<-ctx.Done()
			atomic.AddInt64(&s.connectedF, -1)
		}()
	}
}

func (s *State) max(connType soul.ConnectionType) {
	switch connType {
	case peer.ConnectionType:
		for {
			s.mu.RLock()
			ok := s.connectedP < s.client.config.MaxPeers
			s.mu.RUnlock()

			if ok {
				break
			} else {
				time.Sleep(1 * time.Second)
				continue
			}
		}

	case distributed.ConnectionType:
		for {
			s.mu.RLock()
			ok := len(s.children) < s.client.config.MaxChildren
			s.mu.RUnlock()

			if ok {
				break
			} else {
				time.Sleep(1 * time.Second)
				continue
			}
		}

	case file.ConnectionType:
		for {
			s.mu.RLock()
			ok := s.connectedF < s.client.config.MaxFileConnections
			s.mu.RUnlock()

			if ok {
				break
			} else {
				time.Sleep(1 * time.Second)
				continue
			}
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

	search := s.client.Relays.FileSearch.Listener(1)
	defer search.Close()

	embed := s.client.Relays.EmbeddedMessage.Listener(1)
	defer embed.Close()

	reset := s.client.Relays.ResetDistributed.Listener(1)
	defer reset.Close()

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
				s.log.Debug().Msg("init")

				// Communicate to server that it should not send us more parents.
				s.mu.RLock()
				_, err := server.Write(s.client.Conn(), &server.HaveNoParent{Have: false})
				s.mu.RUnlock()
				if err != nil {
					s.log.Warn().Err(err).Msg("have")
					return
				}

				s.log.Debug().Msg("stop receiving parents from server")

				s.distributed(parents)

				s.log.Debug().Msg("no parent connected, trying again")

				s.mu.RLock()
				_, err = server.Write(s.client.Conn(), &server.HaveNoParent{Have: true})
				s.mu.RUnlock()
				if err != nil {
					s.log.Warn().Err(err).Msg("have")
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

		case search := <-search.Ch():
			go s.responseS(search)

			// We are root in our distributed branch.
		case embed := <-embed.Ch():
			s.log.Debug().Any("embed", embed).Msg("embed")

			go func(embed *server.EmbeddedMessage) {
				s.mu.RLock()
				for _, peer := range s.children {
					conn, _ := peer.Conn(distributed.ConnectionType)
					_, err := distributed.Write(conn, &distributed.EmbeddedMessage{Code: embed.Code, Message: embed.Message})
					if err != nil {
						s.log.Warn().Err(err).Msg("search")
						continue
					}
				}

				s.level = 0

				// Disconnect from all children.
				for _, p := range s.children {
					go p.cancelD()
				}
				s.mu.RUnlock()
			}(embed)

			// Reset the distributed search. We do not need to do anything about the s.distributed() method.
			// It will cancelled by the ctxD.Done() along with all children connections.
		case <-reset.Ch():
			s.log.Debug().Msg("reset")

			s.mu.Lock()
			for _, p := range s.children {
				go p.cancelD()
			}
			s.children = make([]*Peer, 0)
			s.parent = nil
			s.mu.Unlock()
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

			// TODO: finish me.
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

				s.initializers(ctx, init.ConnectionType, p, init.Conn, init.Obfuscated, il)
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
					cl.Warn().Err(err).Msg("dial")
					return
				}

				cl.Debug().Msg("connected to peer")

				_, err = peer.Write(conn, &peer.PierceFirewall{Token: connect.Token}, useObfuscatedPort)
				if err != nil {
					cl.Warn().Err(err).Msg("pierce firewall")
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

				s.initializers(ctx, connect.Type, p, conn, useObfuscatedPort, cl)
			}(connect)
		}
	}
}

func (s *State) initializers(ctx context.Context, connType soul.ConnectionType, p *Peer, conn net.Conn, useObfuscatedPort bool, l zerolog.Logger) {
	switch connType {
	// If the connection is of type P (peer), start the file response listener.
	// The previous fileResponse, if any, is cancelled in the Logic step (or NewPeer)
	// if the connection is of peer P. Thus it is safe to start a new one here.
	case peer.ConnectionType:
		go s.fileResponse(p)

	// If the connection is of type D (distributed), we send the branch root and level
	// to the peer.
	case distributed.ConnectionType:
		s.mu.RLock()
		_, err := distributed.Write(conn, &distributed.BranchRoot{Root: s.root})
		if err != nil {
			l.Warn().Err(err).Msg("branch root")
			return
		}

		_, err = distributed.Write(conn, &distributed.BranchLevel{Level: s.level})
		if err != nil {
			l.Warn().Err(err).Msg("branch level")
			return
		}
		s.mu.RUnlock()
	}

	p.New(connType, conn, useObfuscatedPort)

	l.Debug().Msg("peer updated")

	go s.count(ctx, connType, p)
}

func (s *State) distributed(m *server.PossibleParents) {
	for _, parent := range m.Parents {
		pl := s.log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str("parent", parent.Username).Logger()

		pl.Debug().Msg("trying parent")

		s.mu.Lock()
		for _, v := range s.children {
			go v.cancelD()
		}
		s.children = make([]*Peer, 0)
		s.mu.Unlock()

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", parent.IP.String(), parent.Port))
		if err != nil {
			pl.Warn().Err(err).Msg("distributed")
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
		s.parent = p
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

		_, err = distributed.Write(conn, &peer.PeerInit{
			Username:       s.client.config.Username,
			ConnectionType: distributed.ConnectionType,
		})
		if err != nil {
			pl.Warn().Err(err).Msg("init")
			continue
		}

		pl.Info().Msg("parent connected")

		_, err = server.Write(s.client.Conn(), &server.AcceptChildren{Accept: s.client.config.AcceptChildren})
		if err != nil {
			log.Err(err).Msg("accept children")
			continue
		}

		for {
			select {
			case <-p.ctxD.Done():
				return

			case branch := <-branch.Ch():
				pl.Debug().Any("branch", branch).Msg("branch")
				_, err = server.Write(s.client.Conn(), &server.BranchRoot{Root: branch.Root})
				if err != nil {
					pl.Warn().Err(err).Msg("branch root")
					continue
				}

				s.root = branch.Root

			case level := <-level.Ch():
				pl.Debug().Int32("level", level.Level).Msg("level")

				_, err = server.Write(s.client.Conn(), &server.BranchLevel{Level: int(level.Level + 1)})
				if err != nil {
					pl.Warn().Err(err).Msg("branch level")
					continue
				}

				s.level = level.Level + 1

			// We are first child in our distributed branch.
			case embed := <-embed.Ch():
				pl.Debug().Any("embed", embed).Msg("embed")

				go func(embed *distributed.EmbeddedMessage) {
					s.mu.RLock()
					for _, peer := range s.children {
						switch embed.Code {
						case distributed.CodeSearch:
							message := new(distributed.Search)
							err = message.Deserialize(bytes.NewBuffer(embed.Message))
							if err != nil {
								pl.Warn().Err(err).Msg("search")
								continue
							}

							conn, _ := peer.Conn(distributed.ConnectionType)
							_, err = distributed.Write(conn, message)
							if err != nil {
								pl.Warn().Err(err).Msg("search")
								continue
							}

						case distributed.CodeBranchRoot:
							message := new(distributed.BranchRoot)
							err = message.Deserialize(bytes.NewBuffer(embed.Message))
							if err != nil {
								pl.Warn().Err(err).Msg("root")
								continue
							}

							conn, _ := peer.Conn(distributed.ConnectionType)
							_, err = distributed.Write(conn, message)
							if err != nil {
								pl.Warn().Err(err).Msg("root")
								continue
							}

						case distributed.CodeBranchLevel:
							message := new(distributed.BranchLevel)
							err = message.Deserialize(bytes.NewBuffer(embed.Message))
							if err != nil {
								pl.Warn().Err(err).Msg("level")
								continue
							}

							conn, _ := peer.Conn(distributed.ConnectionType)
							_, err = distributed.Write(conn, message)
							if err != nil {
								pl.Warn().Err(err).Msg("level")
								continue
							}
						}
					}
					s.mu.RUnlock()
				}(embed)

			case search := <-search.Ch():
				go func(search *distributed.Search) {
					s.mu.RLock()
					for _, peer := range s.children {
						conn, _ := peer.Conn(distributed.ConnectionType)
						_, err = distributed.Write(conn, search)
						if err != nil {
							pl.Warn().Err(err).Msg("search")
							continue
						}
					}
					s.mu.RUnlock()
				}(search)

				go s.responseD(search)
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
				for _, f := range fileResponse.Results {
					channel <- &File{
						Username: fileResponse.Username,
						Token:    fileResponse.Token,
						Queue:    fileResponse.Queue,
						File:     &f,
					}
				}

				for _, f := range fileResponse.PrivateResults {
					channel <- &File{
						Username: fileResponse.Username,
						Token:    fileResponse.Token,
						Queue:    fileResponse.Queue,
						File:     &f,
					}
				}

			case false:
				s.log.Debug().Any("message", fileResponse).Msg("search not found")
			}
		}
	}
}

func (s *State) responseS(search *server.FileSearch) {
	s.response(nil, search)
}

// TODO: finish me Byron.
func (s *State) responseD(search *distributed.Search) {
	s.response(search, nil)
}

func (s *State) response(di *distributed.Search, se *server.FileSearch) {
	switch {
	case di != nil:

	case se != nil:

	}
}
