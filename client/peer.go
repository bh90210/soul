package client

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"sync"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/distributed"
	"github.com/bh90210/soul/file"
	"github.com/bh90210/soul/peer"
	"github.com/bh90210/soul/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/teivah/broadcast"
)

// Peer represents a peer.
type Peer struct {
	Relays peerRelays

	username       string
	ip             net.IP
	port           int
	obfuscatedPort int
	status         server.UserStatus
	averageSpeed   int
	queued         int
	privileged     bool

	config        *Config
	mu            sync.RWMutex
	firewallToken soul.Token

	conn       net.Conn
	obfuscated bool
	ctx        context.Context
	cancel     context.CancelFunc
	queue      chan map[peer.Code]io.Reader

	connD   net.Conn
	ctxD    context.Context
	cancelD context.CancelFunc
	queueD  chan map[distributed.Code]io.Reader

	muF   sync.RWMutex
	connF map[soul.Token]net.Conn

	log zerolog.Logger
}

// NewPeer creates a new peer.
func NewPeer(config *Config, message *peer.PeerInit) *Peer {
	p := &Peer{
		username: message.Username,
		config:   config,
		connF:    make(map[soul.Token]net.Conn),
		queue:    make(chan map[peer.Code]io.Reader),
		queueD:   make(chan map[distributed.Code]io.Reader),
	}

	p.log = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	p.log = p.log.Level(config.LogLevel)
	p.log = p.log.With().Str("username", p.username).Logger()

	p.relayInit()

	return p
}

// New is the main logic for the peer.
func (p *Peer) New(connType soul.ConnectionType, conn net.Conn, obfuscated bool) {
	switch connType {
	case peer.ConnectionType:
		p.mu.Lock()
		p.log = log.With().Str("username", p.username).Bool("obfuscated", obfuscated).Logger()

		if p.cancel != nil {
			p.cancel()
		}

		p.conn = conn
		p.ctx, p.cancel = context.WithCancel(context.Background())
		p.obfuscated = obfuscated
		p.mu.Unlock()

		go p.read(p.ctx, obfuscated)
		go p.deserialize(p.ctx)

		go func() {
			<-p.ctx.Done()
			p.conn.Close()
		}()

	case distributed.ConnectionType:
		p.mu.Lock()
		if p.cancelD != nil {
			p.cancelD()
		}

		p.connD = conn
		p.ctxD, p.cancelD = context.WithCancel(context.Background())
		p.mu.Unlock()

		go p.readD(p.ctxD)
		go p.deserializeD(p.ctxD)

		go func() {
			<-p.ctxD.Done()
			p.connD.Close()
		}()

	case file.ConnectionType:
		init := new(file.TransferInit)
		err := init.Deserialize(conn)
		if err != nil && !errors.Is(err, io.EOF) {
			p.log.Warn().Err(err).Msg("transfer init deserialize")
			return
		}

		p.muF.Lock()
		p.connF[init.Token] = conn
		p.muF.Unlock()

		p.log.Info().Msg("file F connection")
	}
}

// Conn returns the connection.
func (p *Peer) Conn(connType soul.ConnectionType, token ...soul.Token) (net.Conn, bool) {
	switch connType {
	case peer.ConnectionType:
		p.mu.RLock()
		defer p.mu.RUnlock()

		return p.conn, p.obfuscated

	case distributed.ConnectionType:
		p.mu.RLock()
		defer p.mu.RUnlock()

		return p.connD, false

	case file.ConnectionType:
		t := token[0]

		p.muF.Lock()
		conn, _ := p.connF[t]
		delete(p.connF, t)
		p.muF.Unlock()

		return conn, false

	default:
		return nil, false
	}
}

func (p *Peer) read(ctx context.Context, obfuscated bool) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			p.mu.RLock()
			r, size, code, err := peer.Read(peer.Code(0), p.conn, obfuscated)
			p.mu.RUnlock()

			if err != nil && !errors.Is(err, io.EOF) {
				if errors.Is(err, net.ErrClosed) {
					p.mu.RLock()
					p.cancel()
					p.mu.RUnlock()
					continue
				}

				p.log.Warn().Err(err).Msg("peer read")
				continue
			}

			// TODO: re-check this solution.
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

func (p *Peer) readD(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			p.mu.RLock()
			r, size, code, err := distributed.Read(p.connD)
			p.mu.RUnlock()

			if err != nil && !errors.Is(err, io.EOF) {
				if errors.Is(err, net.ErrClosed) {
					p.mu.RLock()
					go p.cancelD()
					p.mu.RUnlock()
					return
				}

				p.log.Warn().Err(err).Msg("distributed read")
				continue
			}

			// TODO: some clients cause a flood of empty messages with code 0. Investigate.
			if code == distributed.Code(0) && size == 0 {
				p.mu.RLock()
				go p.cancelD()
				p.mu.RUnlock()
				continue
			}

			p.queueD <- map[distributed.Code]io.Reader{distributed.Code(code): r}
		}
	}
}

func (p *Peer) deserialize(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

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
							p.log.Warn().Err(err).Msg("file search response deserialize")
							continue
						}

						p.Relays.FileSearchResponse.NotifyCtx(ctx, m)

					case peer.CodeFolderContentsRequest:
						m := new(peer.FolderContentsRequest)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("folder contents request deserialize")
							continue
						}

						p.Relays.FolderContentsRequest.NotifyCtx(ctx, m)

					case peer.CodeFolderContentsResponse:
						m := new(peer.FolderContentsResponse)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("folder contents response deserialize")
							continue
						}

						p.Relays.FolderContentsResponse.NotifyCtx(ctx, m)

					case peer.CodePlaceInQueueRequest:
						m := new(peer.PlaceInQueueRequest)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("place in queue request deserialize")
							continue
						}

						p.Relays.PlaceInQueueRequest.NotifyCtx(ctx, m)

					case peer.CodePlaceInQueueResponse:
						m := new(peer.PlaceInQueueResponse)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("place in queue response deserialize")
							continue
						}

						p.Relays.PlaceInQueueResponse.NotifyCtx(ctx, m)

					case peer.CodeQueueUpload:
						m := new(peer.QueueUpload)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("queue upload deserialize")
							continue
						}

						p.Relays.QueueUpload.NotifyCtx(ctx, m)

					case peer.CodeSharedFileListResponse:
						m := new(peer.SharedFileListResponse)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("shared file list response deserialize")
							continue
						}

						p.Relays.SharedFileListResponse.NotifyCtx(ctx, m)

					case peer.CodeSharedFileListRequest:
						m := new(peer.SharedFileListRequest)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("shared file list request deserialize")
							continue
						}

						p.Relays.SharedFileListRequest.NotifyCtx(ctx, m)

					case peer.CodeTransferRequest:
						m := new(peer.TransferRequest)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("transfer request deserialize")
							continue
						}

						p.Relays.TransferRequest.NotifyCtx(ctx, m)

					case peer.CodeTransferResponse:
						m := new(peer.TransferResponse)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("transfer response deserialize")
							continue
						}

						p.Relays.TransferResponse.NotifyCtx(ctx, m)

					case peer.CodeUploadDenied:
						m := new(peer.UploadDenied)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("upload denied deserialize")
							continue
						}

						p.Relays.UploadDenied.NotifyCtx(ctx, m)

					case peer.CodeUploadFailed:
						m := new(peer.UploadFailed)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("upload failed deserialize")
							continue
						}

						p.Relays.UploadFailed.NotifyCtx(ctx, m)

					case peer.CodeUserInfoRequest:
						m := new(peer.UserInfoRequest)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("deserialize")
							continue
						}

						p.Relays.UserInfoRequest.NotifyCtx(ctx, m)

					case peer.CodeUserInfoResponse:
						m := new(peer.UserInfoResponse)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("user info response deserialize")
							continue
						}

						p.Relays.UserInfoResponse.NotifyCtx(ctx, m)
					}
				}
			}(m)
		}
	}
}

func (p *Peer) deserializeD(ctx context.Context) {
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
							p.log.Warn().Err(err).Msg("branch level deserialize")
							continue
						}

						p.Relays.Distributed.BranchLevel.NotifyCtx(ctx, m)

					case distributed.CodeBranchRoot:
						m := new(distributed.BranchRoot)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("branch root deserialize")
							continue
						}

						p.Relays.Distributed.BranchRoot.NotifyCtx(ctx, m)

					case distributed.CodeEmbeddedMessage:
						m := new(distributed.EmbeddedMessage)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("embedded message deserialize")
							continue
						}

						p.Relays.Distributed.EmbeddedMessage.NotifyCtx(ctx, m)

					case distributed.CodeSearch:
						m := new(distributed.Search)
						err := m.Deserialize(r)
						if err != nil {
							p.log.Warn().Err(err).Msg("search deserialize")
							continue
						}

						p.Relays.Distributed.Search.NotifyCtx(ctx, m)
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

	Distributed *distributedRelays
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

	p.Relays.Distributed = new(distributedRelays)
	p.Relays.Distributed.BranchLevel = broadcast.NewRelay[*distributed.BranchLevel]()
	p.Relays.Distributed.BranchRoot = broadcast.NewRelay[*distributed.BranchRoot]()
	p.Relays.Distributed.EmbeddedMessage = broadcast.NewRelay[*distributed.EmbeddedMessage]()
	p.Relays.Distributed.Search = broadcast.NewRelay[*distributed.Search]()
}
