package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/distributed"
	"github.com/bh90210/soul/file"
	"github.com/bh90210/soul/peer"
	"github.com/bh90210/soul/server"
	"github.com/rs/zerolog/log"
	"github.com/teivah/broadcast"
)

// Peer represents a peer.
type Peer struct {
	Relays peerRelays
	Writer chan []byte

	username       string
	ip             net.IP
	port           int
	obfuscatedPort int
	status         server.UserStatus
	averageSpeed   int
	queued         int
	privileged     bool

	ctx     context.Context
	cancel  context.CancelFunc
	ctxD    context.Context
	cancelD context.CancelFunc

	config          *Config
	mu              sync.RWMutex
	firewallToken   soul.Token
	conn            net.Conn
	distributedConn net.Conn
	muF             sync.RWMutex
	fileConns       map[soul.Token]net.Conn

	queue             chan map[peer.Code]io.Reader
	queueD            chan map[distributed.Code]io.Reader
	relaysD           distributedRelays
	distributedWriter chan []byte

	initListeners            *peerInitListeners
	initDistributedListeners *distributedInitListeners
}

// NewPeer creates a new peer.
func NewPeer(config *Config, message *peer.PeerInit, conn net.Conn) *Peer {
	p := &Peer{
		username:  message.RemoteUsername,
		Writer:    make(chan []byte),
		config:    config,
		fileConns: make(map[soul.Token]net.Conn),
		queue:     make(chan map[peer.Code]io.Reader),
		queueD:    make(chan map[distributed.Code]io.Reader),
	}

	p.relayInit()
	p.listenersInit()

	p.Logic(message.ConnectionType, conn)

	return p
}

// Logic is the main logic for the peer.
func (p *Peer) Logic(connType soul.ConnectionType, conn net.Conn) {
	switch connType {
	case peer.ConnectionType:
		p.mu.Lock()
		if p.cancel != nil {
			go p.cancel()
		}

		p.conn = conn
		p.ctx, p.cancel = context.WithCancel(context.Background())
		p.mu.Unlock()

		go p.write(p.ctx)
		go p.read(p.ctx)
		go p.deserialize(p.ctx)

	case distributed.ConnectionType:
		p.mu.Lock()
		if p.cancelD != nil {
			go p.cancelD()
		}

		p.distributedConn = conn
		p.ctxD, p.cancelD = context.WithCancel(context.Background())
		p.mu.Unlock()

		go p.writeD(p.ctxD)
		go p.readD(p.ctxD)

	case file.ConnectionType:
		init := new(file.TransferInit)
		err := init.Deserialize(conn)
		if err != nil {
			log.Error().Err(err).Msg("transfer init deserialize")
			return
		}

		p.muF.Lock()
		p.fileConns[init.Token] = conn
		p.muF.Unlock()

		log.Info().Str("username", p.username).Msg("file F connection")
	}
}

// File accepts a token and returns a connection and a uint64. The connection is the file connection
// and the uint64 is the number of bytes of the file that the peer has previously uploaded.
func (p *Peer) File(ctx context.Context, token soul.Token, offset uint64) (net.Conn, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default:
			p.muF.Lock()
			conn, ok := p.fileConns[token]
			delete(p.fileConns, token)
			p.muF.Unlock()

			if !ok {
				if p.ip != nil {
					log.Info().Str("username", p.username).Str("ip", p.ip.String()).Int("port", p.port).Msg("file F connection not found")

					conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", p.ip.String(), p.port))
					if err != nil {
						return nil, err
					}

					init := new(peer.PeerInit)
					message, err := init.Serialize(p.config.Username, file.ConnectionType)
					if err != nil {
						return nil, err
					}

					_, err = conn.Write(message)
					if err != nil {
						return nil, err
					}

					initT := new(file.TransferInit)
					message, err = initT.Serialize(token)
					if err != nil {
						return nil, err
					}

					_, err = conn.Write(message)
					if err != nil {
						return nil, err
					}

					off := new(file.Offset)
					message, err = off.Serialize(offset)
					if err != nil {
						return nil, err
					}

					_, err = conn.Write(message)
					if err != nil {
						return nil, err
					}

					return conn, nil
				}

				log.Info().Str("username", p.username).Msg("file F connection not found")
				time.Sleep(time.Second)
				continue
			}

			if ok {
				log.Info().Str("username", p.username).Msg("file F connection found")
				o := new(file.Offset)
				message, err := o.Serialize(offset)
				if err != nil {
					return nil, err
				}

				_, err = conn.Write(message)
				if err != nil {
					return nil, err
				}

				return conn, nil
			}
		}
	}
}

func (p *Peer) write(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case m := <-p.Writer:
			p.mu.RLock()
			_, err := peer.MessageWrite(p.conn, m)
			p.mu.RUnlock()
			if err != nil {
				log.Err(err).Str("username", p.username).Msg("peer write")
				continue
			}
		}
	}
}

func (p *Peer) read(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			p.mu.RLock()
			r, size, code, err := peer.MessageRead(peer.Code(0), p.conn)
			p.mu.RUnlock()
			if err != nil && !errors.Is(err, io.EOF) {
				if errors.Is(err, net.ErrClosed) {
					p.mu.RLock()
					p.cancel()
					p.mu.RUnlock()
					continue
				}

				log.Error().Err(err).Str("username", p.username).Msg("peer read")
				continue
			}

			if code == peer.Code(0) && size == 0 {
				p.mu.RLock()
				p.cancel()
				p.mu.RUnlock()
				continue
			}

			p.queue <- map[peer.Code]io.Reader{peer.Code(code): r}
		}
	}
}

func (p *Peer) writeD(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case m := <-p.distributedWriter:
			p.mu.RLock()
			_, err := distributed.MessageWrite(p.distributedConn, m)
			p.mu.RUnlock()
			if err != nil {
				log.Error().Err(err).Str("username", p.username).Msg("distributed write")
				continue
			}
		}
	}
}

func (p *Peer) readD(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			p.mu.RLock()
			r, _, code, err := distributed.MessageRead(p.distributedConn)
			p.mu.RUnlock()
			if err != nil && !errors.Is(err, io.EOF) {
				if errors.Is(err, net.ErrClosed) {
					time.Sleep(time.Second) // TODO: check if this is necessary
					continue
				}

				log.Error().Err(err).Str("username", p.username).Msg("distributed read")
				continue
			}

			log.Debug().Str("username", p.username).Str("code", code.String()).Msg("distributed read")
			p.queueD <- map[distributed.Code]io.Reader{distributed.Code(code): r}
		}
	}
}

func (p *Peer) deserialize(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case m := <-p.queueD:
			go func(m map[distributed.Code]io.Reader) {
				for code, r := range m {
					ctx, cancel := context.WithTimeout(context.Background(), p.config.Timeout)
					defer cancel()

					switch code {
					case distributed.CodeBranchLevel:
						m := new(distributed.BranchLevel)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("branch level deserialize")
							continue
						}

						p.relaysD.BranchLevel.NotifyCtx(ctx, m)

					case distributed.CodeBranchRoot:
						m := new(distributed.BranchRoot)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("branch root deserialize")
							continue
						}

						p.relaysD.BranchRoot.NotifyCtx(ctx, m)

					case distributed.CodeEmbeddedMessage:
						m := new(distributed.EmbeddedMessage)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("embedded message deserialize")
							continue
						}

						p.relaysD.EmbeddedMessage.NotifyCtx(ctx, m)

					case distributed.CodeSearch:
						m := new(distributed.Search)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("search deserialize")
							continue
						}

						p.relaysD.Search.NotifyCtx(ctx, m)
					}
				}
			}(m)

		case m := <-p.queue:
			go func(m map[peer.Code]io.Reader) {
				for code, r := range m {
					ctx, cancel := context.WithTimeout(context.Background(), p.config.Timeout)
					defer cancel()

					switch code {
					case peer.CodeFileSearchResponse:
						m := new(peer.FileSearchResponse)
						err := m.Deserialize(r)
						if err != nil && !errors.Is(err, io.EOF) {
							log.Error().Err(err).Msg("file search response deserialize")
							continue
						}

						p.Relays.FileSearchResponse.NotifyCtx(ctx, m)

					case peer.CodeFolderContentsRequest:
						m := new(peer.FolderContentsRequest)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("folder contents request deserialize")
							continue
						}

						p.Relays.FolderContentsRequest.NotifyCtx(ctx, m)

					case peer.CodeFolderContentsResponse:
						m := new(peer.FolderContentsResponse)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("folder contents response deserialize")
							continue
						}

						p.Relays.FolderContentsResponse.NotifyCtx(ctx, m)

					case peer.CodePlaceInQueueRequest:
						m := new(peer.PlaceInQueueRequest)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("place in queue request deserialize")
							continue
						}

						p.Relays.PlaceInQueueRequest.NotifyCtx(ctx, m)

					case peer.CodePlaceInQueueResponse:
						m := new(peer.PlaceInQueueResponse)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("place in queue response deserialize")
							continue
						}

						p.Relays.PlaceInQueueResponse.NotifyCtx(ctx, m)

					case peer.CodeQueueUpload:
						m := new(peer.QueueUpload)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("queue upload deserialize")
							continue
						}

						p.Relays.QueueUpload.NotifyCtx(ctx, m)

					case peer.CodeSharedFileListResponse:
						m := new(peer.SharedFileListResponse)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("shared file list response deserialize")
							continue
						}

						p.Relays.SharedFileListResponse.NotifyCtx(ctx, m)

					case peer.CodeSharedFileListRequest:
						m := new(peer.SharedFileListRequest)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("shared file list request deserialize")
							continue
						}

						p.Relays.SharedFileListRequest.NotifyCtx(ctx, m)

					case peer.CodeTransferRequest:
						m := new(peer.TransferRequest)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("transfer request deserialize")
							continue
						}

						p.Relays.TransferRequest.NotifyCtx(ctx, m)

					case peer.CodeTransferResponse:
						m := new(peer.TransferResponse)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("transfer response deserialize")
							continue
						}

						p.Relays.TransferResponse.NotifyCtx(ctx, m)

					case peer.CodeUploadDenied:
						m := new(peer.UploadDenied)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("upload denied deserialize")
							continue
						}

						p.Relays.UploadDenied.NotifyCtx(ctx, m)

					case peer.CodeUploadFailed:
						m := new(peer.UploadFailed)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("upload failed deserialize")
							continue
						}

						p.Relays.UploadFailed.NotifyCtx(ctx, m)

					case peer.CodeUserInfoRequest:
						m := new(peer.UserInfoRequest)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("deserialize")
							continue
						}

						p.Relays.UserInfoRequest.NotifyCtx(ctx, m)

					case peer.CodeUserInfoResponse:
						m := new(peer.UserInfoResponse)
						err := m.Deserialize(r)
						if err != nil {
							log.Error().Err(err).Msg("user info response deserialize")
							continue
						}

						p.Relays.UserInfoResponse.NotifyCtx(ctx, m)
					}
				}
			}(m)
		}
	}
}

type peerRelays struct {
	FileSearchResponse     *broadcast.Relay[*peer.FileSearchResponse]
	FolderContentsRequest  *broadcast.Relay[*peer.FolderContentsRequest]
	FolderContentsResponse *broadcast.Relay[*peer.FolderContentsResponse]
	PlaceInQueueRequest    *broadcast.Relay[*peer.PlaceInQueueRequest]
	PlaceInQueueResponse   *broadcast.Relay[*peer.PlaceInQueueResponse]
	QueueUpload            *broadcast.Relay[*peer.QueueUpload]
	SharedFileListResponse *broadcast.Relay[*peer.SharedFileListResponse]
	SharedFileListRequest  *broadcast.Relay[*peer.SharedFileListRequest]
	TransferRequest        *broadcast.Relay[*peer.TransferRequest]
	TransferResponse       *broadcast.Relay[*peer.TransferResponse]
	UploadDenied           *broadcast.Relay[*peer.UploadDenied]
	UploadFailed           *broadcast.Relay[*peer.UploadFailed]
	UserInfoRequest        *broadcast.Relay[*peer.UserInfoRequest]
	UserInfoResponse       *broadcast.Relay[*peer.UserInfoResponse]
}

type distributedRelays struct {
	BranchLevel     *broadcast.Relay[*distributed.BranchLevel]
	BranchRoot      *broadcast.Relay[*distributed.BranchRoot]
	EmbeddedMessage *broadcast.Relay[*distributed.EmbeddedMessage]
	Search          *broadcast.Relay[*distributed.Search]
}

func (p *Peer) relayInit() {
	p.Relays.FileSearchResponse = broadcast.NewRelay[*peer.FileSearchResponse]()
	p.Relays.FolderContentsRequest = broadcast.NewRelay[*peer.FolderContentsRequest]()
	p.Relays.FolderContentsResponse = broadcast.NewRelay[*peer.FolderContentsResponse]()
	p.Relays.PlaceInQueueRequest = broadcast.NewRelay[*peer.PlaceInQueueRequest]()
	p.Relays.PlaceInQueueResponse = broadcast.NewRelay[*peer.PlaceInQueueResponse]()
	p.Relays.QueueUpload = broadcast.NewRelay[*peer.QueueUpload]()
	p.Relays.SharedFileListResponse = broadcast.NewRelay[*peer.SharedFileListResponse]()
	p.Relays.SharedFileListRequest = broadcast.NewRelay[*peer.SharedFileListRequest]()
	p.Relays.TransferRequest = broadcast.NewRelay[*peer.TransferRequest]()
	p.Relays.TransferResponse = broadcast.NewRelay[*peer.TransferResponse]()
	p.Relays.UploadDenied = broadcast.NewRelay[*peer.UploadDenied]()
	p.Relays.UploadFailed = broadcast.NewRelay[*peer.UploadFailed]()
	p.Relays.UserInfoRequest = broadcast.NewRelay[*peer.UserInfoRequest]()
	p.Relays.UserInfoResponse = broadcast.NewRelay[*peer.UserInfoResponse]()

	p.relaysD.BranchLevel = broadcast.NewRelay[*distributed.BranchLevel]()
	p.relaysD.BranchRoot = broadcast.NewRelay[*distributed.BranchRoot]()
	p.relaysD.EmbeddedMessage = broadcast.NewRelay[*distributed.EmbeddedMessage]()
	p.relaysD.Search = broadcast.NewRelay[*distributed.Search]()
}

type peerInitListeners struct {
	fileSearchResponse     <-chan *peer.FileSearchResponse
	folderContentsRequest  <-chan *peer.FolderContentsRequest
	folderContentsResponse <-chan *peer.FolderContentsResponse
	placeInQueueRequest    <-chan *peer.PlaceInQueueRequest
	placeInQueueResponse   <-chan *peer.PlaceInQueueResponse
	queueUpload            <-chan *peer.QueueUpload
	sharedFileListResponse <-chan *peer.SharedFileListResponse
	sharedFileListRequest  <-chan *peer.SharedFileListRequest
	transferRequest        <-chan *peer.TransferRequest
	transferResponse       <-chan *peer.TransferResponse
	uploadDenied           <-chan *peer.UploadDenied
	uploadFailed           <-chan *peer.UploadFailed
	userInfoRequest        <-chan *peer.UserInfoRequest
	userInfoResponse       <-chan *peer.UserInfoResponse
}

type distributedInitListeners struct {
	branchLevel     <-chan *distributed.BranchLevel
	branchRoot      <-chan *distributed.BranchRoot
	embeddedMessage <-chan *distributed.EmbeddedMessage
	search          <-chan *distributed.Search
}

func (p *Peer) listenersInit() {
	p.initListeners = &peerInitListeners{
		fileSearchResponse:     p.Relays.FileSearchResponse.Listener(1).Ch(),
		folderContentsRequest:  p.Relays.FolderContentsRequest.Listener(1).Ch(),
		folderContentsResponse: p.Relays.FolderContentsResponse.Listener(1).Ch(),
		placeInQueueRequest:    p.Relays.PlaceInQueueRequest.Listener(1).Ch(),
		placeInQueueResponse:   p.Relays.PlaceInQueueResponse.Listener(1).Ch(),
		queueUpload:            p.Relays.QueueUpload.Listener(1).Ch(),
		sharedFileListResponse: p.Relays.SharedFileListResponse.Listener(1).Ch(),
		sharedFileListRequest:  p.Relays.SharedFileListRequest.Listener(1).Ch(),
		transferRequest:        p.Relays.TransferRequest.Listener(1).Ch(),
		transferResponse:       p.Relays.TransferResponse.Listener(1).Ch(),
		uploadDenied:           p.Relays.UploadDenied.Listener(1).Ch(),
		uploadFailed:           p.Relays.UploadFailed.Listener(1).Ch(),
		userInfoRequest:        p.Relays.UserInfoRequest.Listener(1).Ch(),
		userInfoResponse:       p.Relays.UserInfoResponse.Listener(1).Ch(),
	}

	p.initDistributedListeners = &distributedInitListeners{
		branchLevel:     p.relaysD.BranchLevel.Listener(1).Ch(),
		branchRoot:      p.relaysD.BranchRoot.Listener(1).Ch(),
		embeddedMessage: p.relaysD.EmbeddedMessage.Listener(1).Ch(),
		search:          p.relaysD.Search.Listener(1).Ch(),
	}
}
