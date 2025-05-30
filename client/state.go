package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"slices"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/distributed"
	"github.com/bh90210/soul/file"
	"github.com/bh90210/soul/peer"
	"github.com/bh90210/soul/server"
	"github.com/charlievieth/fastwalk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// HundredKb 100Kb is the size of the buffer for file downloads.
const HundredKb = 100000

// QueueUpload is the message sent to the upload queue.
type QueueUpload struct {
	Filename string
	Peer     *Peer
}

// State represents the client state.
type State struct {
	Incoming chan *Search

	client               *Client
	searches             map[soul.Token]chan *File
	peers                map[string]*Peer // TODO: Periodically empty.
	addToQueue           chan *QueueUpload
	queuePositionRequest chan *queuePositionRequest
	queueSizeRequest     chan chan int
	mu                   sync.RWMutex

	connectedP int64
	connectedF int64

	level    int32
	root     string
	parent   *Peer
	children []*Peer

	shared *peer.SharedFileListResponse

	log zerolog.Logger
}

// NewState returns a new State.
func NewState(c *Client) *State {
	s := &State{
		Incoming:             make(chan *Search),
		client:               c,
		searches:             make(map[soul.Token]chan *File),
		peers:                make(map[string]*Peer),
		addToQueue:           make(chan *QueueUpload),
		queuePositionRequest: make(chan *queuePositionRequest),
		queueSizeRequest:     make(chan chan int),
		shared:               &peer.SharedFileListResponse{},
	}

	s.log = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	s.log = s.log.Level(c.config.LogLevel)

	// Load library directory.
	shared := make(map[string][]peer.File, 0)
	var mu sync.Mutex
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			return nil // returning the error stops iteration
		}

		if !d.IsDir() {
			i, err := d.Info()
			if err != nil {
				return err
			}

			mu.Lock()
			shared[filepath.Dir(path)] = append(shared[filepath.Dir(path)], peer.File{
				Name:       path,
				Size:       uint64(i.Size()),
				Extension:  filepath.Ext(d.Name()),
				Attributes: []peer.Attribute{},
			})
			mu.Unlock()
		}

		return nil
	}

	err := fastwalk.Walk(&fastwalk.DefaultConfig, c.config.Library, walkFn)
	if err != nil {
		s.log.Fatal().Err(err).Msg("walk")
	}

	for k, v := range shared {
		s.shared.Directories = append(s.shared.Directories, peer.Directory{
			Name:  k,
			Files: v,
		})
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
	if s.client.config.OwnPortObfuscated != 0 {
		port.ObfuscatedPort = s.client.config.OwnPortObfuscated
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

	// Once we are logged in to the server, start processing incoming messages from server and peers.
	go s.peer(ctx)
	go s.server(ctx)
	go s.queue(ctx)

	return nil
}

// File represents a file to be downloaded.
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

// Status represents the status of a file transfer.
type Status string

const (
	// StatusQueued indicates that the file is queued for download.
	StatusQueued Status = "queued"
	// StatusStarting indicates that the file is starting to download.
	StatusStarting Status = "starting"
	// StatusReceived indicates that the file has been received.
	StatusReceived Status = "received"
)

// ErrNoPeer is returned when the peer is not found.
var ErrNoPeer = errors.New("no peer")

// Download sends download message to the server and listens for the responses.
func (s *State) Download(ctx context.Context, f *File) (status chan string, e chan error) {
	// Try find the the username of the file to download among the peers.
	s.mu.RLock()
	p, found := s.peers[f.Username]
	s.mu.RUnlock()

	status = make(chan string, 10)
	e = make(chan error, 1)

	// If the peer is not found, return an error.
	if !found {
		e <- ErrNoPeer
		return
	}

	go s.download(ctx, p, f, status, e)

	return
}

func (s *State) download(ctx context.Context, p *Peer, f *File, status chan string, e chan error) {
	// Init peer listeners relating to the file transfer.
	tRequest := p.Relays.TransferRequest.Listener(1)
	defer tRequest.Close()

	failed := p.Relays.UploadFailed.Listener(1)
	defer failed.Close()

	denied := p.Relays.UploadDenied.Listener(1)
	defer denied.Close()

	placeInQueue := p.Relays.PlaceInQueueResponse.Listener(1)
	defer p.Relays.PlaceInQueueResponse.Close()

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
				if m != nil {
					e <- m.Reason
					return
				}
			}
		}
	}()

	sl := s.log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str(fmt.Sprintf("%v", f.Token), f.Name).Logger()

	sl.Debug().Msg("queue upload")

	// Send the peer the queue upload message.
	conn, obfuscated := p.Conn(peer.ConnectionType)
	if conn == nil {
		sl.Warn().Msg("no connection")
		e <- errors.New("no connection")
		return
	}

	_, err := peer.Write(conn, &peer.QueueUpload{Filename: f.Name}, obfuscated)
	if err != nil {
		e <- err
		return
	}

	status <- string(StatusQueued)

	// Send a place queue request.
	_, err = peer.Write(conn, &peer.PlaceInQueueRequest{Filename: f.Name}, obfuscated)

	sl.Debug().Msg("waiting transfer request")

	// When peer is ready to start the file transfer, it sends a transfer request.
	var transfer *peer.TransferRequest
	for {
		var transferRequest bool

		select {
		case <-ctx.Done():
			sl.Warn().Msg("context done")
			return

		case piq := <-placeInQueue.Ch():
			sl.Debug().Msg("place in queue")
			if piq.Filename != f.Name {
				continue
			}

			status <- fmt.Sprint(piq.Place)

		case transfer = <-tRequest.Ch():
			if transfer.Filename != f.Name {
				continue
			}

			transferRequest = true
		}

		if transferRequest {
			sl.Debug().Msg("transfer request")
			break
		}
	}

	sl.Debug().Msg("transfer response")

	// We reply to the transfer request with a transfer response.
	_, err = peer.Write(conn, &peer.TransferResponse{
		Token:   transfer.Token,
		Allowed: true,
	}, obfuscated)
	if err != nil {
		e <- err
		return
	}

	filepath := path.Join(s.client.config.DownloadFolder, path.Base(f.Name))

	sl.Debug().Str("path", filepath).Msg("transfer response sent")

	// Stat for the destination file.
	info, err := os.Stat(filepath)
	if err != nil {
		if !os.IsNotExist(err) {
			e <- err
			return
		}
	}

	var localFile *os.File
	// If file does not exist, create it and pass 0 to the offset.
	if os.IsNotExist(err) {
		sl.Debug().Msg("file does not exist")

		localFile, err = os.Create(filepath)
		if err != nil {
			sl.Debug().Msg(err.Error())
			e <- err
			return
		}

		defer localFile.Close()

		info, err = localFile.Stat()
		if err != nil {
			e <- err
			return
		}

	} else {
		// If file exists count the length and pass it to the offset.
		sl.Debug().Msg("file exists")

		localFile, err = os.OpenFile(filepath, os.O_RDWR, 0644)
		if err != nil {
			e <- err
			return
		}

		defer localFile.Close()

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

	s.log.Debug().Any("info", info).Msg("file info")

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
			defer fileConn.Close()
			break
		}

		sl.Debug().Msg("waiting for file connection")
	}

	sl.Debug().Int64("offset", info.Size()).Msg("sending offset")

	_, err = file.Write(fileConn, &file.Offset{Offset: uint64(info.Size())})
	if err != nil {
		e <- err
		return
	}

	sl.Debug().Msg("offset sent")

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

		if readSoFar == int64(f.Size) {
			break
		}

		status <- fmt.Sprintf("%v%%", readSoFar*100/int64(f.Size))
	}

	sl.Debug().Msg("CopyN exited")

	status <- string(StatusReceived)

	return
}

// ErrNoFiles is returned when no files are found.
var ErrNoFiles = errors.New("no files")

// ErrNoUsername is returned when no username is found.
var ErrNoUsername = errors.New("no username")

// ErrNoToken is returned when no token is found.
var ErrNoToken = errors.New("no token")

// Respond sends a response to the search request.
func (s *State) Respond(ctx context.Context, files []*File) error {
	s.log.Debug().Any("files", files).Msg("responding")

	if len(files) == 0 {
		return ErrNoFiles
	}

	username := files[0].Username
	if username == "" {
		return ErrNoUsername
	}

	token := files[0].Token
	if token == 0 {
		return ErrNoToken
	}

	var results []peer.File
	for _, f := range files {
		results = append(results, *f.File)
	}

	conn, obfuscated, err := s.connect(ctx, username, token)
	if err != nil {
		s.log.Warn().Err(err).Msg("respond connect")
		return err
	}

	// Peers may or may not have an active peer type connection at the time of this request.
	// So we need to account for that. We will be trying to reconnect until context is done or
	// an unrecoverable error occurs.
	for {
		select {
		case <-ctx.Done():
			return errors.New("context done")

		default:
			// Try sending the response to peer.
			_, err := peer.Write(conn, &peer.FileSearchResponse{
				Username: s.client.config.Username,
				Token:    token,
				Results:  results,
				FreeSlot: true,
				Queue:    0,
			}, obfuscated)
			if err != nil {
				return fmt.Errorf("file search response: %w", err)
			}

			return nil
		}
	}
}

// TODO: polish it (duplicate code etc.)
func (s *State) connect(ctx context.Context, username string, token soul.Token) (net.Conn, bool, error) {
	// Search for peer.
	s.mu.RLock()
	p, ok := s.peers[username]
	s.mu.RUnlock()

	// If the peer is not found try to connect.
	if !ok {
		// Open a GetPeerAddress listener.
		gpa := s.client.Relays.GetPeerAddress.Listener(1)
		defer gpa.Close()

		// Let the server know the username we need the address for.
		_, err := server.Write(s.client.Conn(), &server.GetPeerAddress{Username: username})
		if err != nil {
			return nil, false, err
		}

		// The listener may receive multiple addresses,
		// so we need to find the one that matches the username.
		var address *server.GetPeerAddress
		for {
			var found bool
			select {
			case <-ctx.Done():
				return nil, false, errors.New("context done")

			case a := <-gpa.Ch():
				if a.Username != username {
					continue
				}

				address = a
				found = true
				break
			}

			if found {
				break
			}
		}

		var port int
		if address.ObfuscatedPort != 0 {
			port = address.ObfuscatedPort
		} else {
			port = address.Port
		}

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", address.IP.String(), port))
		if err != nil {
			s.log.Warn().Err(err).Msg("respond direct connection")

			for {
				_, err = server.Write(s.client.Conn(), &server.ConnectToPeer{
					Token:    token,
					Username: username,
					Type:     peer.ConnectionType,
				})
				if err != nil {
					s.log.Warn().Err(err).Msg("respond indirect connection")
					return nil, false, err
				}

				s.log.Debug().Msg("waiting for peer to connect")

				var connected bool
				select {
				case <-ctx.Done():
					return nil, false, errors.New("context done")

				default:
					s.mu.RLock()
					p, ok = s.peers[username]
					s.mu.RUnlock()

					if ok {
						connected = true
						break
					}
				}

				if connected {
					break
				}

				time.Sleep(5 * time.Second)
			}
		} else {
			s.mu.Lock()
			p = NewPeer(s.client.config, &peer.PeerInit{
				Username:       username,
				ConnectionType: peer.ConnectionType,
			})

			p.ip = address.IP
			p.port = address.Port
			p.obfuscatedPort = address.ObfuscatedPort

			s.peers[p.username] = p
			s.mu.Unlock()

			s.initializers(ctx, peer.ConnectionType, p, conn, address.ObfuscatedPort != 0, s.log)

			_, err = peer.Write(conn, &peer.PeerInit{
				Username:       s.client.config.Username,
				ConnectionType: peer.ConnectionType,
			}, address.ObfuscatedPort != 0)
			if err != nil {
				return nil, false, err
			}
		}
	}

	// TODO: finish checking if connection is active and retrying for {}.
	conn, obfuscated := p.Conn(peer.ConnectionType)
	// This should not happen at this stage.
	// Nevertheless, we need to check if the connection is nil.
	// TODO: if nil try reconnecting to peer.
	if conn == nil {
		return nil, false, errors.New("connection nill")
	}

	ui := p.Relays.UserInfoResponse.Listener(1)
	defer ui.Close()

	for {
		s.log.Info().Msg("waiting for user info response")
		_, err := peer.Write(conn, &peer.UserInfoRequest{}, obfuscated)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				s.log.Warn().Msg("connection closed, trying to reconnect")
				// We try to directly reconnect to peer.
				// Open a GetPeerAddress listener.
				gpa := s.client.Relays.GetPeerAddress.Listener(1)
				defer gpa.Close()

				// Let the server know the username we need the address for.
				_, err := server.Write(s.client.Conn(), &server.GetPeerAddress{Username: username})
				if err != nil {
					return nil, false, err
				}

				// The listener may receive multiple addresses,
				// so we need to find the one that matches the username.
				var address *server.GetPeerAddress
				for {
					s.log.Info().Msg("waiting for peer address")

					var found bool
					select {
					case <-ctx.Done():
						return nil, false, errors.New("context done")

					case a := <-gpa.Ch():
						if a.Username != username {
							continue
						}

						address = a
						found = true
						break
					}

					if found {
						break
					}
				}

				var port int
				if address.ObfuscatedPort != 0 {
					port = address.ObfuscatedPort
				} else {
					port = address.Port
				}

				s.log.Debug().Msg("peer address found")

				conn, err = net.Dial("tcp", fmt.Sprintf("%s:%v", address.IP.String(), port))
				if err != nil {
					s.log.Warn().Err(err).Msg("respond direct connection")
					return nil, false, err
				}

				s.initializers(ctx, peer.ConnectionType, p, conn, address.ObfuscatedPort != 0, s.log)

				_, err = peer.Write(conn, &peer.PeerInit{
					Username:       s.client.config.Username,
					ConnectionType: peer.ConnectionType,
				}, address.ObfuscatedPort != 0)
				if err != nil {
					return nil, false, err
				}
			} else {
				return nil, false, fmt.Errorf("file search response: %w", err)
			}
		} else {
			break
		}
	}

	s.log.Debug().Msg("user info response sent, now waiting for response")

	select {
	case <-ctx.Done():
		return nil, false, errors.New("context done")

	case <-ui.Ch():
		s.log.Debug().Msg("user info response received")
	}

	return conn, obfuscated, nil
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
		case firewall := <-s.client.Firewall:
			s.log.Debug().Any("firewall", firewall).Msg("firewall")

		// Peer directly connects to us.
		case init := <-s.client.Init:
			go func(init *PeerInit) {
				if init.Username == s.client.config.Username {
					s.log.Debug().Msg("can't connect to self")
					return
				}

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
				if connect.Username == s.client.config.Username {
					s.log.Debug().Msg("can't connect to self")
					return
				}

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

	// TODO: Relogged code 41. ParentMinSpeed code 83. ParentSpeedRatio code 84.

	for {
		select {
		case <-ctx.Done():
			return

		case status := <-statusListener.Ch():
			s.mu.Lock()
			defer s.mu.Unlock()

			p, ok := s.peers[status.Username]
			if ok {
				p.status = status.Status
				p.privileged = status.Privileged
			} else {
				s.log.Warn().Str("status", status.Status.String()).Str("username", status.Username).Msg("peer not found")
			}

		case stats := <-statsListener.Ch():
			s.mu.Lock()
			defer s.mu.Unlock()

			p, ok := s.peers[stats.Username]
			if ok {
				p.averageSpeed = stats.Speed
				p.queued = stats.Uploads
			} else {
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
			defer s.mu.Unlock()

			p, ok := s.peers[watch.Username]
			if ok {
				p.status = watch.Status
				p.averageSpeed = watch.AverageSpeed
				p.queued = watch.UploadNumber
			} else {
				s.log.Warn().Any("watch", watch).Msg("peer not found")
			}

		case search := <-search.Ch():
			// We do not want to respond to our own search.
			if search.Username == s.client.config.Username {
				continue
			}

			go s.serverSearch(search)

			// We are root in our distributed branch.
		case embed := <-embed.Ch():
			s.log.Debug().Any("embed", embed).Msg("embed")

			go func(embed *server.EmbeddedMessage) {
				s.mu.RLock()
				for _, peer := range s.children {
					conn, _ := peer.Conn(distributed.ConnectionType)
					if conn == nil {
						s.log.Warn().Msg("no connection")
						continue
					}

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

func (s *State) queue(ctx context.Context) {
	pl := s.log.With().Str("process", "upload queue").Logger()

	var queue []*QueueUpload
	var mu sync.RWMutex

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case que := <-s.addToQueue:
				pl.Debug().Any("queue", que).Msg("add to queue")

				var alreadyInQueue bool
				mu.Lock()
				for _, v := range queue {
					if v.Filename == que.Filename && v.Peer.username == que.Peer.username {
						alreadyInQueue = true
						break
					}
				}

				if !alreadyInQueue {
					queue = append(queue, que)
				}
				mu.Unlock()

			case piq := <-s.queuePositionRequest:
				var position int
				mu.RLock()
				for i, file := range queue {
					if file.Filename == piq.filename && file.Peer.username == piq.username {
						position = i + 1
						break
					}
				}
				mu.RUnlock()

				if position == 0 {
					pl.Warn().Any("place in queue", piq).Msg("queue position request")
					continue
				}

				err := piq.response(position)
				if err != nil {
					pl.Warn().Any("place in queue", piq).Err(err).Msg("queue position request")
					continue
				}

			case replyChannel := <-s.queueSizeRequest:
				replyChannel <- len(queue)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return

		default:
			mu.RLock()
			noQueue := len(queue) == 0
			mu.RUnlock()

			if noQueue {
				time.Sleep(1 * time.Second)
				continue
			}

			// Check if we reached the max number of file connections.
			s.max(file.ConnectionType)

			// Pop the first file from the upload queue.
			mu.Lock()
			que := queue[0]
			queue = queue[1:]
			mu.Unlock()

			go func(que *QueueUpload) {
				token := soul.NewToken()

				ul := s.log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str("file", que.Filename).Str("peer", que.Peer.username).Uint32("token", uint32(token)).Logger()

				transferResponse := que.Peer.Relays.TransferResponse.Listener(0)
				defer transferResponse.Close()

				localFile, err := os.OpenFile(que.Filename, os.O_RDONLY, 0644)
				if err != nil {
					ul.Warn().Err(err).Msg("open file")
					return
				}

				defer localFile.Close()

				info, err := localFile.Stat()
				if err != nil {
					ul.Warn().Err(err).Msg("stat file")
					return
				}

				conn, obfuscated := que.Peer.Conn(peer.ConnectionType)
				if conn == nil { // TODO: finish me.
				}

				_, err = peer.Write(conn, &peer.TransferRequest{
					Direction: peer.UploadToPeer,
					Token:     token,
					Filename:  info.Name(),
					FileSize:  uint64(info.Size()),
				}, obfuscated)
				if err != nil {
					ul.Warn().Err(err).Msg("transfer request")
					return
				}

				var tResponse *peer.TransferResponse
				for {
					var moveOn bool

					select {
					case <-ctx.Done():
						return

					case tResponse = <-transferResponse.Ch():
						if tResponse.Token != token {
							continue
						}

						if !tResponse.Allowed {
							ul.Warn().Err(tResponse.Reason).Msg("transfer response")
							return
						}

						moveOn = true
					}

					if moveOn {
						break
					}
				}

				s.log.Debug().Any("response", tResponse).Msg("response")

				conn, err = net.Dial("tcp", fmt.Sprintf("%s:%v", que.Peer.ip.String(), que.Peer.port))
				if err != nil {
					ul.Warn().Err(err).Msg("dial")
					return
				}

				_, err = peer.Write(conn, &peer.PeerInit{
					Username:       s.client.config.Username,
					ConnectionType: file.ConnectionType,
				}, false)
				if err != nil {
					ul.Warn().Err(err).Msg("peer init")
					return
				}

				_, err = file.Write(conn, &file.TransferInit{Token: token})
				if err != nil {
					ul.Warn().Err(err).Msg("transfer init")
					return
				}

				s.log.Debug().Msg("transfer init")

				offset := new(file.Offset)
				err = offset.Deserialize(conn)
				if err != nil {
					ul.Warn().Err(err).Msg("offset")
					return
				}

				s.log.Debug().Any("off", offset).Msg("offset")

				// Send the file.
				n, err := localFile.Seek(int64(offset.Offset), io.SeekCurrent)
				if err != nil {
					ul.Warn().Err(err).Msg("seek")
					return
				}

				if n != int64(offset.Offset) {
					ul.Warn().Int64("n", n).Uint64("offset", offset.Offset).Msg("seek not equal")
					return
				}

				s.log.Debug().Uint64("offset", offset.Offset).Int64("file size", info.Size()).Msg("sending file")

				_, err = io.CopyN(conn, localFile, info.Size())
				if err != nil && !errors.Is(err, io.EOF) {
					ul.Warn().Err(err).Msg("copy")
					return
				}

				conn.Close()
			}(que)
		}
	}
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

				r := branch.Root

				s.mu.Lock()
				s.root = r
				s.mu.Unlock()

				_, err = server.Write(s.client.Conn(), &server.BranchRoot{Root: r})
				if err != nil {
					pl.Warn().Err(err).Msg("branch root")
					continue
				}

			case level := <-level.Ch():
				pl.Debug().Int32("level", level.Level).Msg("level")

				l := level.Level + 1

				s.mu.Lock()
				s.level = l
				s.mu.Unlock()

				_, err = server.Write(s.client.Conn(), &server.BranchLevel{Level: int(l)})
				if err != nil {
					pl.Warn().Err(err).Msg("branch level")
					continue
				}

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
							if conn == nil {
								pl.Warn().Msg("no connection")
								continue
							}

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
							if conn == nil {
								pl.Warn().Msg("no connection")
								continue
							}

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
							if conn == nil {
								pl.Warn().Msg("no connection")
								continue
							}

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
						if conn == nil {
							pl.Warn().Msg("no connection")
							continue
						}

						_, err = distributed.Write(conn, search)
						if err != nil {
							pl.Warn().Err(err).Msg("search")
							continue
						}
					}
					s.mu.RUnlock()
				}(search)

				go s.distributedSearch(search)
			}
		}
	}
}

func (s *State) peerRequests(ctx context.Context, p *Peer, wg *sync.WaitGroup) {
	prl := s.log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Str("username", p.username).Logger()

	fileSearch := p.Relays.FileSearchResponse.Listener(1)
	defer fileSearch.Close()

	qu := p.Relays.QueueUpload.Listener(1)
	defer qu.Close()

	sfl := p.Relays.SharedFileListRequest.Listener(1)
	defer sfl.Close()

	ui := p.Relays.UserInfoRequest.Listener(1)
	defer ui.Close()

	fc := p.Relays.FolderContentsRequest.Listener(1)
	defer fc.Close()

	piq := p.Relays.PlaceInQueueRequest.Listener(1)
	defer piq.Close()

	if wg != nil {
		wg.Done()
	}

	for {
		select {
		case <-ctx.Done():
			return

		case fileResponse := <-fileSearch.Ch():
			prl.Debug().Any("fileResponse", fileResponse).Msg("file search response")

			go func(fileResponse *peer.FileSearchResponse) {
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
			}(fileResponse)

		case qu := <-qu.Ch():
			prl.Debug().Any("qu", qu).Msg("queue upload request")

			s.addToQueue <- &QueueUpload{
				Filename: qu.Filename,
				Peer:     p,
			}

			prl.Debug().Any("qu", qu).Msg("queue upload request sent")

		case <-sfl.Ch():
			prl.Debug().Msg("shared file list request")

			conn, obfuscated := p.Conn(peer.ConnectionType)
			if conn == nil {
				prl.Warn().Msg("no connection")
				continue
			}

			s.mu.RLock()
			_, err := peer.Write(conn, s.shared, obfuscated)
			s.mu.RUnlock()
			if err != nil {
				prl.Warn().Err(err).Msg("shared file list response")
				continue
			}

			prl.Debug().Msg("shared file list response sent")

		case <-ui.Ch():
			prl.Debug().Msg("user info request")

			queueSize := make(chan int)
			s.queueSizeRequest <- queueSize
			size := <-queueSize

			conn, obfuscated := p.Conn(peer.ConnectionType)
			if conn == nil {
				prl.Warn().Msg("no connection")
				continue
			}

			_, err := peer.Write(conn, &peer.UserInfoResponse{
				Description: s.client.config.Description,
				Picture:     s.client.config.Picture,
				TotalUpload: uint32(s.client.config.MaxFileConnections),
				QueueSize:   uint32(size),
				FreeSlots:   s.client.config.MaxFileConnections > s.connectedF,
			}, obfuscated)
			if err != nil {
				prl.Warn().Err(err).Msg("user info response")
				continue
			}

			prl.Debug().Msg("user info response sent")

		case fc := <-fc.Ch():
			prl.Debug().Any("fc", fc).Msg("folder contents request")

			conn, obfuscated := p.Conn(peer.ConnectionType)
			if conn == nil {
				prl.Warn().Any("fc", fc).Msg("no connection")
				continue
			}

			var folders []peer.Directory
			s.mu.RLock()
			for _, directory := range s.shared.Directories {
				if strings.Contains(directory.Name, fc.Folder) {
					folders = append(folders, peer.Directory{
						Name:  directory.Name,
						Files: directory.Files,
					})
				}
			}
			s.mu.RUnlock()

			_, err := peer.Write(conn, &peer.FolderContentsResponse{
				Token:   fc.Token,
				Folder:  fc.Folder,
				Folders: folders,
			}, obfuscated)
			if err != nil {
				prl.Warn().Err(err).Any("fc", fc).Msg("folder contents response")
				continue
			}

			prl.Debug().Any("fc", fc).Msg("folder contents response sent")

		case piq := <-piq.Ch():
			prl.Debug().Any("piq", piq).Msg("place in queue request")

			go func(piq *peer.PlaceInQueueRequest, p *Peer) {
				s.queuePositionRequest <- &queuePositionRequest{
					username: p.username,
					filename: piq.Filename,
					response: func(place int) error {
						s.mu.RLock()
						defer s.mu.RUnlock()

						conn, obfuscated := p.Conn(peer.ConnectionType)
						if conn == nil {
							return errors.New("connection nill")
						}

						_, err := peer.Write(conn, &peer.PlaceInQueueResponse{
							Filename: piq.Filename,
							Place:    uint32(place),
						}, obfuscated)
						if err != nil {
							prl.Warn().Err(err).Any("piq", piq).Msg("place in queue response")
							return err
						}

						return nil
					},
				}
			}(piq, p)
		}
	}
}

type queuePositionRequest struct {
	username string
	filename string
	response func(place int) error
}

func (s *State) initializers(ctx context.Context, connType soul.ConnectionType, p *Peer, conn net.Conn, useObfuscatedPort bool, l zerolog.Logger) {
	wg, ctx := p.New(connType, conn, useObfuscatedPort)

	switch connType {
	// If the connection is of type P (peer), start the peer listeners.
	case peer.ConnectionType:
		go s.peerRequests(ctx, p, wg)

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

	l.Debug().Msg("peer updated")

	go s.count(ctx, connType, p)
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
			s.log.Debug().Int64("connectedF", s.connectedF).Int64("max", s.client.config.MaxFileConnections).Msg("max")
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

func (s *State) serverSearch(search *server.FileSearch) {
	s.request(&Search{
		Username: search.Username,
		Token:    search.Token,
		Query:    search.SearchQuery,
	})
}

// TODO: finish me Byron.
func (s *State) distributedSearch(search *distributed.Search) {
	s.request(&Search{
		Username: search.Username,
		Token:    search.Token,
		Query:    search.Query,
	})
}

// Search is the search request sent to the client.
type Search struct {
	Username string
	Token    soul.Token
	Query    string
}

func (s *State) request(r *Search) {
	if r != nil {
		s.Incoming <- r
	}
}
