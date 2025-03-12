package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/bh90210/soul/peer"
	"github.com/bh90210/soul/server"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/teivah/broadcast"
)

// Client is a minimal SoulSeek client that can handle login, search uploads and downloads,
// plus a tiny chat bot like functionality.
type Client struct {
	// Relays contains all possible server messages the client will deserialize and notify.
	Relays serverRelays
	Writer chan []byte
	// firewallTokens is a map of incoming Firewall tokens and connections.
	Firewall chan *PierceFirewall
	// Init initiates a connection with a peer. Consumers will receive a deserialized peer.PeerInit message.
	Init chan *PeerInit

	config *Config
	mu     sync.RWMutex
	// SoulSeek network connection.
	conn net.Conn
	// listener net.Listener for incoming peer connections.
	listener net.Listener
	// queue is filled by reading the conn. It is used to deserialize messages.
	// It is a channel of maps where the key is the server message code and the value is the message.
	queue chan map[server.Code]io.Reader

	dialling bool
	wg       sync.WaitGroup
	cancel   context.CancelFunc

	zerolog.Logger
}

type PierceFirewall struct {
	*peer.PierceFirewall
	Conn net.Conn
}

type PeerInit struct {
	*peer.PeerInit
	Conn net.Conn
}

type Config struct {
	SoulSeekAddress string
	SoulSeekPort    int
	OwnPort         int
	Username        string
	Password        string
	SharedFolders   int
	SharedFiles     int
	LogLevel        zerolog.Level // TODO: fix logger.
	MaxUploads      int
	Timeout         time.Duration
	LoginTimeout    time.Duration
	DownloadFolder  string
	MaxPeers        int64
}

func DefaultConfig() *Config {
	return &Config{
		SoulSeekAddress: "server.slsknet.org",
		SoulSeekPort:    2242,
		OwnPort:         2234,
		Username:        gonanoid.MustGenerate("soulseek", 7),
		Password:        gonanoid.MustGenerate("0123456789qwertyuiop", 10),
		SharedFolders:   1,
		SharedFiles:     1,
		LogLevel:        zerolog.Disabled,
		MaxUploads:      10,
		Timeout:         2 * time.Second,
		LoginTimeout:    1 * time.Second,
		DownloadFolder:  os.TempDir(),
		MaxPeers:        100,
	}
}

// New connects to the server. It uses the values in the Config.
func New(conf ...*Config) (*Client, error) {
	c := &Client{
		config: DefaultConfig(),
	}

	if len(conf) > 0 {
		c.config = conf[0]
	}

	c.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	c.Logger = c.Logger.With().Str("username", c.config.Username).Logger()

	// Init all necessary maps and channels.
	c.Writer = make(chan []byte)
	c.queue = make(chan map[server.Code]io.Reader)
	c.Firewall = make(chan *PierceFirewall)
	c.Init = make(chan *PeerInit)

	c.relaysInit()

	go c.deserialize()

	return c, nil
}

func (c *Client) Dial(ctx context.Context, cancel context.CancelFunc) error {
	c.mu.RLock()
	if c.cancel != nil {
		go c.cancel()
	}
	c.mu.RUnlock()

	err := c.dial()
	if err != nil {
		go cancel()
		return err
	}

	// Listen for incoming peer connections.
	c.mu.Lock()
	if c.listener != nil {
		err = c.listener.Close()
		if err != nil {
			c.mu.Unlock()
			go cancel()
			return err
		}
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", c.config.OwnPort))
	if err != nil {
		c.mu.Unlock()
		go cancel()
		return err
	}

	c.listener = listener
	c.cancel = cancel
	c.mu.Unlock()

	go c.read(ctx)
	go c.write(ctx)
	go c.listen(ctx)

	go func() {
		<-ctx.Done()
		c.close()
	}()

	return nil
}

// close the c.conn connection to server.
func (c *Client) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// dial starts a new connection with the server.
func (c *Client) dial() error {
	// We must lock the c.dialling variable to prevent multiple dialling attempts.
	c.mu.Lock()
	switch c.dialling {
	// If dialling is true we wait for the other process to finish.
	case true:
		c.Debug().Msg("waiting for dialling to finish")
		// We must unlock the mutex before waiting for the other process to finish.
		c.mu.Unlock()
		// All parallel dialling processes will wait here for the first one to finish.
		c.wg.Wait()
		return nil

	// If dialling is false we start the dialling process.
	case false:
		c.Debug().Msg("dialling to server")

		if c.conn != nil {
			// If there is an existing connection we close it.
			c.Debug().Msg("closing existing connection")
			c.conn.Close()
		}

		// We set dialling true while in the mutex lock.
		c.dialling = true
		// We add a new dialling process to the wait group while in the mutex lock.
		c.wg.Add(1)
		c.mu.Unlock()

		defer c.wg.Done()
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", c.config.SoulSeekAddress, c.config.SoulSeekPort))
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.dialling = false
	c.mu.Unlock()

	c.Debug().Msg("dialling to server successful once")

	return nil
}

// Write sends a message to the c.conn connection to the server.
func (c *Client) write(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case m := <-c.Writer:
			c.mu.RLock()
			_, err := server.MessageWrite(c.conn, m)
			c.mu.RUnlock()
			if err != nil {
				c.Err(err).Msg("server write")
				continue
			}
		}
	}
}

// read messages from the server continuously.
func (c *Client) read(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			c.mu.RLock()
			r, _, code, err := server.MessageRead(c.conn)
			c.mu.RUnlock()
			if err != nil {
				c.Err(err).Msg("server read")
				continue
			}

			// Send the message to the deserialization queue.
			c.queue <- map[server.Code]io.Reader{code: r}
		}
	}
}

// TODO: implement limits (see config's limits.)
func (c *Client) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			c.mu.RLock()
			conn, err := c.listener.Accept()
			c.mu.RUnlock()
			if err != nil {
				c.Error().Err(err).Msg("accept TCP")
				continue
			}

			go func(conn net.Conn) {
				// Upon a new connection we reed for init codes.
				r, _, code, err := peer.MessageRead(peer.CodeInit(0), conn)
				if err != nil {
					c.Error().Err(err).Msg("accept TCP")
					return
				}

				switch code {
				// We receive this message as a response to a ConnectToPeer message we sent to the server
				// while trying to directly connect to a peer (PeerInit 1) but failed to do so.
				case peer.CodePierceFirewall:
					firewall := new(peer.PierceFirewall)
					err = firewall.Deserialize(r)
					if err != nil {
						c.Error().Err(err).Msg("pierce firewall")
						return
					}

					// Consumers must check the ConnectToPeer messages and match the token.
					// Next they must wait for the PeerInit message on this connection.
					c.Firewall <- &PierceFirewall{PierceFirewall: firewall, Conn: conn}

					c.Debug().Int("token", int(firewall.Token)).Msg("incoming firewall token")

				// We receive this message when a peer is trying a direct connection to us.
				case peer.CodePeerInit:
					peerInit := new(peer.PeerInit)
					err := peerInit.Deserialize(r)
					if err != nil {
						c.Error().Err(err).Msg("peer init")
						return
					}

					c.Init <- &PeerInit{PeerInit: peerInit, Conn: conn}

					c.Debug().Str("username", peerInit.RemoteUsername).Str("connection type", string(peerInit.ConnectionType)).Msg("peer init")
				}
			}(conn)
		}
	}
}

// deserialize reads messages from the deserialization queue. If successful, it
// sends a notification to all potential listeners with a cancellation context.
func (c *Client) deserialize() {
	// TODO: implement limits, see config.
	for {
		for code, r := range <-c.queue {
			go func(code server.Code, r io.Reader) {
				ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
				defer cancel()

				switch code {
				case server.CodeAdminMessage:
					m := new(server.AdminMessage)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("admin message deserialize")
						return
					}

					c.Relays.AdminMessage.NotifyCtx(ctx, m)

				case server.CodeCantConnectToPeer:
					m := new(server.CantConnectToPeer)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("cant connect to peer deserialize")
						return
					}

					c.Relays.CantConnectToPeer.NotifyCtx(ctx, m)

				case server.CodeCantCreateRoom:
					m := new(server.CantCreateRoom)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("cant create room deserialize")
						return
					}

					c.Relays.CantCreateRoom.NotifyCtx(ctx, m)

				case server.CodeChangePassword:
					m := new(server.ChangePassword)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("change password deserialize")
						return
					}

					c.Relays.ChangePassword.NotifyCtx(ctx, m)

				case server.CodeCheckPrivileges:
					m := new(server.CheckPrivileges)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("check privileges deserialize")
						return
					}

					c.Relays.CheckPrivileges.NotifyCtx(ctx, m)

				case server.CodeConnectToPeer:
					m := new(server.ConnectToPeer)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("connect to peer deserialize")
						return
					}

					c.Relays.ConnectToPeer.NotifyCtx(ctx, m)

				case server.CodeEmbeddedMessage:
					m := new(server.EmbeddedMessage)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("embedded message deserialize")
						return
					}

					c.Relays.EmbeddedMessage.NotifyCtx(ctx, m)

				case server.CodeExcludedSearchPhrases:
					m := new(server.ExcludedSearchPhrases)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("excluded search phrases deserialize")
						return
					}

					c.Relays.ExcludedSearchPhrases.NotifyCtx(ctx, m)

				case server.CodeFileSearch:
					m := new(server.FileSearch)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("file search deserialize")
						return
					}

					c.Relays.FileSearch.NotifyCtx(ctx, m)

				case server.CodeGetPeerAddress:
					m := new(server.GetPeerAddress)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("get peer address deserialize")
						return
					}

					c.Relays.GetPeerAddress.NotifyCtx(ctx, m)

				case server.CodeGetUserStats:
					m := new(server.GetUserStats)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("get user stats deserialize")
						return
					}

					c.Relays.GetUserStats.NotifyCtx(ctx, m)

				case server.CodeGetUserStatus:
					m := new(server.GetUserStatus)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("get user status deserialize")
						return
					}

					c.Relays.GetUserStatus.NotifyCtx(ctx, m)

				case server.CodeJoinRoom:
					m := new(server.JoinRoom)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("join room deserialize")
						return
					}

					c.Relays.JoinRoom.NotifyCtx(ctx, m)

				case server.CodeLeaveRoom:
					m := new(server.LeaveRoom)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("leave room deserialize")
						return
					}

					c.Relays.LeaveRoom.NotifyCtx(ctx, m)

				case server.CodeLogin:
					m := new(server.Login)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("login deserialize")
						return
					}

					c.Relays.Login.NotifyCtx(ctx, m)

				case server.CodeMessageUser:
					m := new(server.MessageUser)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("message user deserialize")
						return
					}

					c.Relays.MessageUser.NotifyCtx(ctx, m)

				case server.CodeParentMinSpeed:
					m := new(server.ParentMinSpeed)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("parent min speed deserialize")
						return
					}

					c.Relays.ParentMinSpeed.NotifyCtx(ctx, m)

				case server.CodeParentSpeedRatio:
					m := new(server.ParentSpeedRatio)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("parent speed ratio deserialize")
						return
					}

					c.Relays.ParentSpeedRatio.NotifyCtx(ctx, m)

				case server.CodePossibleParents:
					m := new(server.PossibleParents)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("possible parents deserialize")
						return
					}

					c.Relays.PossibleParents.NotifyCtx(ctx, m)

				case server.CodePrivateRoomAddOperator:
					m := new(server.PrivateRoomAddOperator)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room add operator deserialize")
						return
					}

					c.Relays.PrivateRoomAddOperator.NotifyCtx(ctx, m)

				case server.CodePrivateRoomAddUser:
					m := new(server.PrivateRoomAddUser)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room add user deserialize")
						return
					}

					c.Relays.PrivateRoomAddUser.NotifyCtx(ctx, m)

				case server.CodePrivateRoomAdded:
					m := new(server.PrivateRoomAdded)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room added deserialize")
						return
					}

					c.Relays.PrivateRoomAdded.NotifyCtx(ctx, m)

				case server.CodePrivateRoomOperatorAdded:
					m := new(server.PrivateRoomOperatorAdded)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room operator added deserialize")
						return
					}

					c.Relays.PrivateRoomOperatorAdded.NotifyCtx(ctx, m)

				case server.CodePrivateRoomOperatorRemoved:
					m := new(server.PrivateRoomOperatorRemoved)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room operator removed deserialize")
						return
					}

					c.Relays.PrivateRoomOperatorRemoved.NotifyCtx(ctx, m)

				case server.CodePrivateRoomOperators:
					m := new(server.PrivateRoomOperators)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room operators deserialize")
						return
					}

					c.Relays.PrivateRoomOperators.NotifyCtx(ctx, m)

				case server.CodePrivateRoomRemoveOperator:
					m := new(server.PrivateRoomRemoveOperator)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room remove operator deserialize")
						return
					}

					c.Relays.PrivateRoomRemoveOperator.NotifyCtx(ctx, m)

				case server.CodePrivateRoomRemoveUser:
					m := new(server.PrivateRoomRemoveUser)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room remove user deserialize")
						return
					}

					c.Relays.PrivateRoomRemoveUser.NotifyCtx(ctx, m)

				case server.CodePrivateRoomRemoved:
					m := new(server.PrivateRoomRemoved)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room removed deserialize")
						return
					}

					c.Relays.PrivateRoomRemoved.NotifyCtx(ctx, m)

				case server.CodePrivateRoomToggle:
					m := new(server.PrivateRoomToggle)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room toggle deserialize")
						return
					}

					c.Relays.PrivateRoomToggle.NotifyCtx(ctx, m)

				case server.CodePrivateRoomUsers:
					m := new(server.PrivateRoomUsers)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("private room users deserialize")
						return
					}

					c.Relays.PrivateRoomUsers.NotifyCtx(ctx, m)

				case server.CodePrivilegedUsers:
					m := new(server.PrivilegedUsers)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("privileged users deserialize")
						return
					}

					c.Relays.PrivilegedUsers.NotifyCtx(ctx, m)

				case server.CodeRelogged:
					m := new(server.Relogged)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("relogged deserialize")
						return
					}

					c.Relays.Relogged.NotifyCtx(ctx, m)

				case server.CodeResetDistributed:
					m := new(server.ResetDistributed)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("reset distributed deserialize")
						return
					}

					c.Relays.ResetDistributed.NotifyCtx(ctx, m)

				case server.CodeRoomList:
					m := new(server.RoomList)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("room list deserialize")
						return
					}

					c.Relays.RoomList.NotifyCtx(ctx, m)

				case server.CodeRoomTicker:
					m := new(server.RoomTicker)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("room ticker deserialize")
						return
					}

					c.Relays.RoomTicker.NotifyCtx(ctx, m)

				case server.CodeRoomTickerAdd:
					m := new(server.RoomTickerAdd)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("room ticker add deserialize")
						return
					}

					c.Relays.RoomTickerAdd.NotifyCtx(ctx, m)

				case server.CodeRoomTickerRemove:
					m := new(server.RoomTickerRemove)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("room ticker remove deserialize")
						return
					}

					c.Relays.RoomTickerRemove.NotifyCtx(ctx, m)

				case server.CodeSayChatroom:
					m := new(server.SayChatroom)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("say chatroom deserialize")
						return
					}

					c.Relays.SayChatroom.NotifyCtx(ctx, m)

				case server.CodeUserJoinedRoom:
					m := new(server.UserJoinedRoom)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("user joined room deserialize")
						return
					}

					c.Relays.UserJoinedRoom.NotifyCtx(ctx, m)

				case server.CodeWatchUser:
					m := new(server.WatchUser)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("watch user deserialize")
						return
					}

					c.Relays.WatchUser.NotifyCtx(ctx, m)

				case server.CodeWishlistInterval:
					m := new(server.WishlistInterval)
					err := m.Deserialize(r)
					if err != nil && !errors.Is(err, io.EOF) {
						c.Err(err).Msg("wishlist interval deserialize")
						return
					}

					c.Relays.WishlistInterval.NotifyCtx(ctx, m)

				default:
					c.Warn().Msgf("message code with no deserialization: %d", code)
				}
			}(code, r)
		}
	}
}

type serverRelays struct {
	AdminMessage               *broadcast.Relay[*server.AdminMessage]
	CantConnectToPeer          *broadcast.Relay[*server.CantConnectToPeer]
	CantCreateRoom             *broadcast.Relay[*server.CantCreateRoom]
	ChangePassword             *broadcast.Relay[*server.ChangePassword]
	CheckPrivileges            *broadcast.Relay[*server.CheckPrivileges]
	ConnectToPeer              *broadcast.Relay[*server.ConnectToPeer]
	EmbeddedMessage            *broadcast.Relay[*server.EmbeddedMessage]
	ExcludedSearchPhrases      *broadcast.Relay[*server.ExcludedSearchPhrases]
	FileSearch                 *broadcast.Relay[*server.FileSearch]
	GetPeerAddress             *broadcast.Relay[*server.GetPeerAddress]
	GetUserStats               *broadcast.Relay[*server.GetUserStats]
	GetUserStatus              *broadcast.Relay[*server.GetUserStatus]
	JoinRoom                   *broadcast.Relay[*server.JoinRoom]
	LeaveRoom                  *broadcast.Relay[*server.LeaveRoom]
	Login                      *broadcast.Relay[*server.Login]
	MessageUser                *broadcast.Relay[*server.MessageUser]
	ParentMinSpeed             *broadcast.Relay[*server.ParentMinSpeed]
	ParentSpeedRatio           *broadcast.Relay[*server.ParentSpeedRatio]
	PossibleParents            *broadcast.Relay[*server.PossibleParents]
	PrivateRoomAddOperator     *broadcast.Relay[*server.PrivateRoomAddOperator]
	PrivateRoomAddUser         *broadcast.Relay[*server.PrivateRoomAddUser]
	PrivateRoomAdded           *broadcast.Relay[*server.PrivateRoomAdded]
	PrivateRoomOperatorAdded   *broadcast.Relay[*server.PrivateRoomOperatorAdded]
	PrivateRoomOperatorRemoved *broadcast.Relay[*server.PrivateRoomOperatorRemoved]
	PrivateRoomOperators       *broadcast.Relay[*server.PrivateRoomOperators]
	PrivateRoomRemoveOperator  *broadcast.Relay[*server.PrivateRoomRemoveOperator]
	PrivateRoomRemoveUser      *broadcast.Relay[*server.PrivateRoomRemoveUser]
	PrivateRoomRemoved         *broadcast.Relay[*server.PrivateRoomRemoved]
	PrivateRoomToggle          *broadcast.Relay[*server.PrivateRoomToggle]
	PrivateRoomUsers           *broadcast.Relay[*server.PrivateRoomUsers]
	PrivilegedUsers            *broadcast.Relay[*server.PrivilegedUsers]
	Relogged                   *broadcast.Relay[*server.Relogged]
	ResetDistributed           *broadcast.Relay[*server.ResetDistributed]
	RoomList                   *broadcast.Relay[*server.RoomList]
	RoomSearch                 *broadcast.Relay[*server.RoomSearch]
	RoomTicker                 *broadcast.Relay[*server.RoomTicker]
	RoomTickerAdd              *broadcast.Relay[*server.RoomTickerAdd]
	RoomTickerRemove           *broadcast.Relay[*server.RoomTickerRemove]
	SayChatroom                *broadcast.Relay[*server.SayChatroom]
	UserJoinedRoom             *broadcast.Relay[*server.UserJoinedRoom]
	WatchUser                  *broadcast.Relay[*server.WatchUser]
	WishlistInterval           *broadcast.Relay[*server.WishlistInterval]
}

func (c *Client) relaysInit() {
	c.Relays.AdminMessage = broadcast.NewRelay[*server.AdminMessage]()
	c.Relays.CantConnectToPeer = broadcast.NewRelay[*server.CantConnectToPeer]()
	c.Relays.CantCreateRoom = broadcast.NewRelay[*server.CantCreateRoom]()
	c.Relays.ChangePassword = broadcast.NewRelay[*server.ChangePassword]()
	c.Relays.CheckPrivileges = broadcast.NewRelay[*server.CheckPrivileges]()
	c.Relays.ConnectToPeer = broadcast.NewRelay[*server.ConnectToPeer]()
	c.Relays.EmbeddedMessage = broadcast.NewRelay[*server.EmbeddedMessage]()
	c.Relays.ExcludedSearchPhrases = broadcast.NewRelay[*server.ExcludedSearchPhrases]()
	c.Relays.FileSearch = broadcast.NewRelay[*server.FileSearch]()
	c.Relays.GetPeerAddress = broadcast.NewRelay[*server.GetPeerAddress]()
	c.Relays.GetUserStats = broadcast.NewRelay[*server.GetUserStats]()
	c.Relays.GetUserStatus = broadcast.NewRelay[*server.GetUserStatus]()
	c.Relays.JoinRoom = broadcast.NewRelay[*server.JoinRoom]()
	c.Relays.LeaveRoom = broadcast.NewRelay[*server.LeaveRoom]()
	c.Relays.Login = broadcast.NewRelay[*server.Login]()
	c.Relays.MessageUser = broadcast.NewRelay[*server.MessageUser]()
	c.Relays.ParentMinSpeed = broadcast.NewRelay[*server.ParentMinSpeed]()
	c.Relays.ParentSpeedRatio = broadcast.NewRelay[*server.ParentSpeedRatio]()
	c.Relays.PossibleParents = broadcast.NewRelay[*server.PossibleParents]()
	c.Relays.PrivateRoomAddOperator = broadcast.NewRelay[*server.PrivateRoomAddOperator]()
	c.Relays.PrivateRoomAddUser = broadcast.NewRelay[*server.PrivateRoomAddUser]()
	c.Relays.PrivateRoomAdded = broadcast.NewRelay[*server.PrivateRoomAdded]()
	c.Relays.PrivateRoomOperatorAdded = broadcast.NewRelay[*server.PrivateRoomOperatorAdded]()
	c.Relays.PrivateRoomOperatorRemoved = broadcast.NewRelay[*server.PrivateRoomOperatorRemoved]()
	c.Relays.PrivateRoomOperators = broadcast.NewRelay[*server.PrivateRoomOperators]()
	c.Relays.PrivateRoomRemoveOperator = broadcast.NewRelay[*server.PrivateRoomRemoveOperator]()
	c.Relays.PrivateRoomRemoveUser = broadcast.NewRelay[*server.PrivateRoomRemoveUser]()
	c.Relays.PrivateRoomRemoved = broadcast.NewRelay[*server.PrivateRoomRemoved]()
	c.Relays.PrivateRoomToggle = broadcast.NewRelay[*server.PrivateRoomToggle]()
	c.Relays.PrivateRoomUsers = broadcast.NewRelay[*server.PrivateRoomUsers]()
	c.Relays.PrivilegedUsers = broadcast.NewRelay[*server.PrivilegedUsers]()
	c.Relays.Relogged = broadcast.NewRelay[*server.Relogged]()
	c.Relays.ResetDistributed = broadcast.NewRelay[*server.ResetDistributed]()
	c.Relays.RoomList = broadcast.NewRelay[*server.RoomList]()
	c.Relays.RoomSearch = broadcast.NewRelay[*server.RoomSearch]()
	c.Relays.RoomTicker = broadcast.NewRelay[*server.RoomTicker]()
	c.Relays.RoomTickerAdd = broadcast.NewRelay[*server.RoomTickerAdd]()
	c.Relays.RoomTickerRemove = broadcast.NewRelay[*server.RoomTickerRemove]()
	c.Relays.SayChatroom = broadcast.NewRelay[*server.SayChatroom]()
	c.Relays.UserJoinedRoom = broadcast.NewRelay[*server.UserJoinedRoom]()
	c.Relays.WatchUser = broadcast.NewRelay[*server.WatchUser]()
	c.Relays.WishlistInterval = broadcast.NewRelay[*server.WishlistInterval]()
}
