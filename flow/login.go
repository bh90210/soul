package flow

import (
	"errors"
	"io"
	"time"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/server"
	"github.com/rs/zerolog/log"
)

type LoginMessage struct {
	Login                 *server.Login
	RoomList              *server.RoomList
	ParentMinSpeed        *server.ParentMinSpeed
	ParentSpeedRatio      *server.ParentSpeedRatio
	WishlistInterval      *server.WishlistInterval
	PrivilegedUsers       *server.PrivilegedUsers
	ExcludedSearchPhrases *server.ExcludedSearchPhrases
	CheckPrivileges       *server.CheckPrivileges
	WatchUser             *server.WatchUser
}

func (c *Client) Login() (*LoginMessage, error) {
	err := c.sendUsernamePassword()
	if err != nil {
		return nil, err
	}

	l := new(LoginMessage)

	err = c.checkUsernamePassword(l)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	err = c.sendRestOfLoginMessages()
	if err != nil {
		return nil, err
	}

	err = c.checkRestOfLoginMessages(l)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return l, nil
}

func (c *Client) sendUsernamePassword() error {
	login := new(server.Login)
	loginMessage, err := login.Serialize(c.Config.Username, c.Config.Password)
	if err != nil {
		return err
	}

	_, err = c.Write(loginMessage)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) checkUsernamePassword(l *LoginMessage) (err error) {
	for {
		c.mu.Lock()
		if r, ok := c.m[server.LoginCode]; ok {
			if len(r) > 0 {
				l.Login = new(server.Login)
				err = l.Login.Deserialize(r[0])
				if err != nil && !errors.Is(err, io.EOF) {
					c.mu.Unlock()
					return
				}

				log.Debug().Any("code", server.LoginCode).Str("greet", l.Login.Greet).IPAddr("ip", l.Login.IP).Str("sum", l.Login.Sum).Send()

				c.m[server.LoginCode] = c.m[server.LoginCode][1:]
				c.mu.Unlock()
				return
			}
		}
		c.mu.Unlock()

		time.Sleep(50 * time.Millisecond)
	}
}

func (c *Client) sendRestOfLoginMessages() error {
	privileges := new(server.CheckPrivileges)
	privilegesMessage, err := privileges.Serialize()
	if err != nil {
		return err
	}

	_, err = c.Write(privilegesMessage)
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	port := new(server.SetListenPort)
	portMessage, err := port.Serialize(uint32(c.Config.SoulseekPort))
	if err != nil {
		return err
	}

	_, err = c.Write(portMessage)
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	status := new(server.SetStatus)
	statusMessage, err := status.Serialize(server.Online)
	if err != nil {
		return err
	}

	_, err = c.Write(statusMessage)
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	shared := new(server.SharedFoldersFiles)
	sharedMessage, err := shared.Serialize(c.Config.SharedFolders, c.Config.SharedFiles)
	if err != nil {
		return err
	}

	_, err = c.Write(sharedMessage)
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	watch := new(server.WatchUser)
	watchMessage, err := watch.Serialize(c.Config.Username)
	if err != nil {
		return err
	}

	_, err = c.Write(watchMessage)
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	noParent := new(server.HaveNoParent)
	parentSearchMessage, err := noParent.Serialize(true)
	if err != nil {
		return err
	}

	_, err = c.Write(parentSearchMessage)
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	root := new(server.BranchRoot)
	rootMessage, err := root.Serialize(c.Config.Username)
	if err != nil {
		return err
	}

	_, err = c.Write(rootMessage)
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	level := new(server.BranchLevel)
	levelMessage, err := level.Serialize(0)
	if err != nil {
		return err
	}

	_, err = c.Write(levelMessage)
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	accept := new(server.AcceptChildren)
	acceptMessage, err := accept.Serialize(true)
	if err != nil {
		return err
	}

	_, err = c.Write(acceptMessage)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) checkRestOfLoginMessages(l *LoginMessage) (err error) {
	necessaryCodes := map[soul.ServerCode]bool{
		server.RoomListCode:              false,
		server.ParentMinSpeedCode:        false,
		server.ParentSpeedRatioCode:      false,
		server.WishlistIntervalCode:      false,
		server.PrivilegedUsersCode:       false,
		server.ExcludedSearchPhrasesCode: false,
		server.CheckPrivilegesCode:       false,
		server.WatchUserCode:             false,
	}

	t := time.Now()

	for {
		if err != nil && !errors.Is(err, io.EOF) {
			return
		}

		// If all necessary codes are present we can return.
		var trueCodes int
		for _, ok := range necessaryCodes {
			if ok {
				trueCodes++
			}
		}
		if trueCodes == len(necessaryCodes) {
			return
		}

		// We give two seconds, after that if all other codes are present
		// we can assume that the CheckPrivileges code will not become present
		// and we continue without it, assuming user has no privileges.
		if time.Since(t) > 2*time.Second {
			if necessaryCodes[server.RoomListCode] &&
				necessaryCodes[server.ParentMinSpeedCode] &&
				necessaryCodes[server.ParentSpeedRatioCode] &&
				necessaryCodes[server.WishlistIntervalCode] &&
				necessaryCodes[server.PrivilegedUsersCode] &&
				necessaryCodes[server.ExcludedSearchPhrasesCode] {
				if !necessaryCodes[server.CheckPrivilegesCode] {
					necessaryCodes[server.CheckPrivilegesCode] = true
					continue
				}

				if !necessaryCodes[server.WatchUserCode] {
					necessaryCodes[server.WatchUserCode] = true
					continue
				}
			}
		}

		for code := range necessaryCodes {
			go func(code soul.ServerCode) {
				c.mu.Lock()
				if r, ok := c.m[code]; ok {
					if len(r) >= 1 {
						switch code {
						case server.RoomListCode:
							l.RoomList = new(server.RoomList)
							err = l.RoomList.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								c.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("rooms", len(l.RoomList.Rooms)).Send()

						case server.ParentMinSpeedCode:
							l.ParentMinSpeed = new(server.ParentMinSpeed)
							err = l.ParentMinSpeed.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								c.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("speed", l.ParentMinSpeed.MinSpeed).Send()

						case server.ParentSpeedRatioCode:
							l.ParentSpeedRatio = new(server.ParentSpeedRatio)
							err = l.ParentSpeedRatio.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								c.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("ratio", l.ParentSpeedRatio.SpeedRatio).Send()

						case server.WishlistIntervalCode:
							l.WishlistInterval = new(server.WishlistInterval)
							err = l.WishlistInterval.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								c.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("interval", l.WishlistInterval.Interval).Send()

						case server.PrivilegedUsersCode:
							l.PrivilegedUsers = new(server.PrivilegedUsers)
							err = l.PrivilegedUsers.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								c.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("users", len(l.PrivilegedUsers.Users)).Send()

						case server.ExcludedSearchPhrasesCode:
							l.ExcludedSearchPhrases = new(server.ExcludedSearchPhrases)
							err = l.ExcludedSearchPhrases.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								c.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("phrases", len(l.ExcludedSearchPhrases.Phrases)).Send()

						case server.CheckPrivilegesCode:
							l.CheckPrivileges = new(server.CheckPrivileges)
							err = l.CheckPrivileges.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								c.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("timeleft", l.CheckPrivileges.TimeLeft).Send()

						case server.WatchUserCode:
							l.WatchUser = new(server.WatchUser)
							err = l.WatchUser.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								c.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Str("user", l.WatchUser.Username).Str("status", l.WatchUser.Status.String()).Send()
						}

						c.m[code] = c.m[code][1:]
						necessaryCodes[code] = true
					}
				}
				c.mu.Unlock()
			}(code)
		}

		time.Sleep(200 * time.Millisecond)
	}
}
