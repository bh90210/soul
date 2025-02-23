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
}

func (s *Server) Login() (*LoginMessage, error) {
	err := s.sendUsernamePassword()
	if err != nil {
		return nil, err
	}

	l := new(LoginMessage)

	err = s.checkUsernamePassword(l)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	err = s.sendRestOfLoginMessages()
	if err != nil {
		return nil, err
	}

	err = s.checkRestOfLoginMessages(l)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return l, nil
}

func (s *Server) sendUsernamePassword() error {
	login := new(server.Login)
	loginMessage, err := login.Serialize(s.Config.Username, s.Config.Password)
	if err != nil {
		return err
	}

	_, err = s.Write(loginMessage)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) checkUsernamePassword(l *LoginMessage) (err error) {
	for {
		s.mu.Lock()
		if r, ok := s.m[server.LoginCode]; ok {
			if len(r) > 0 {
				l.Login = new(server.Login)
				err = l.Login.Deserialize(r[0])
				if err != nil && !errors.Is(err, io.EOF) {
					s.mu.Unlock()
					return
				}

				log.Debug().Any("code", server.LoginCode).Str("greet", l.Login.Greet).IPAddr("ip", l.Login.IP).Str("sum", l.Login.Sum).Msg("login")

				s.m[server.LoginCode] = s.m[server.LoginCode][1:]
				s.mu.Unlock()
				return
			}
		}
		s.mu.Unlock()

		time.Sleep(50 * time.Millisecond)
	}
}

func (s *Server) sendRestOfLoginMessages() error {
	privileges := new(server.CheckPrivileges)
	privilegesMessage, err := privileges.Serialize()
	if err != nil {
		return err
	}

	_, err = s.Write(privilegesMessage)
	if err != nil {
		return err
	}

	port := new(server.SetListenPort)
	portMessage, err := port.Serialize(uint32(s.Config.SoulseekPort))
	if err != nil {
		return err
	}

	_, err = s.Write(portMessage)
	if err != nil {
		return err
	}

	status := new(server.SetStatus)
	statusMessage, err := status.Serialize(server.Online)
	if err != nil {
		return err
	}

	_, err = s.Write(statusMessage)
	if err != nil {
		return err
	}

	shared := new(server.SharedFoldersFiles)
	sharedMessage, err := shared.Serialize(s.Config.SharedFolders, s.Config.SharedFiles)
	if err != nil {
		return err
	}

	_, err = s.Write(sharedMessage)
	if err != nil {
		return err
	}

	watch := new(server.WatchUser)
	watchMessage, err := watch.Serialize(s.Config.Username)
	if err != nil {
		return err
	}

	_, err = s.Write(watchMessage)
	if err != nil {
		return err
	}

	noParent := new(server.HaveNoParent)
	parentSearchMessage, err := noParent.Serialize(true)
	if err != nil {
		return err
	}

	_, err = s.Write(parentSearchMessage)
	if err != nil {
		return err
	}

	root := new(server.BranchRoot)
	rootMessage, err := root.Serialize(s.Config.Username)
	if err != nil {
		return err
	}

	_, err = s.Write(rootMessage)
	if err != nil {
		return err
	}

	level := new(server.BranchLevel)
	levelMessage, err := level.Serialize(0)
	if err != nil {
		return err
	}

	_, err = s.Write(levelMessage)
	if err != nil {
		return err
	}

	accept := new(server.AcceptChildren)
	acceptMessage, err := accept.Serialize(true)
	if err != nil {
		return err
	}

	_, err = s.Write(acceptMessage)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) checkRestOfLoginMessages(l *LoginMessage) (err error) {
	necessaryCodes := map[soul.ServerCode]bool{
		server.RoomListCode:              false,
		server.ParentMinSpeedCode:        false,
		server.ParentSpeedRatioCode:      false,
		server.WishlistIntervalCode:      false,
		server.PrivilegedUsersCode:       false,
		server.ExcludedSearchPhrasesCode: false,
		server.CheckPrivilegesCode:       false,
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
			}
		}

		for code := range necessaryCodes {
			go func(code soul.ServerCode) {
				s.mu.Lock()
				if r, ok := s.m[code]; ok {
					if len(r) > 0 {
						switch code {
						case server.RoomListCode:
							l.RoomList = new(server.RoomList)
							err = l.RoomList.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								s.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("rooms", len(l.RoomList.Rooms)).Msg("roomlist")

						case server.ParentMinSpeedCode:
							l.ParentMinSpeed = new(server.ParentMinSpeed)
							err = l.ParentMinSpeed.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								s.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("speed", l.ParentMinSpeed.MinSpeed).Msg("parentminspeed")

						case server.ParentSpeedRatioCode:
							l.ParentSpeedRatio = new(server.ParentSpeedRatio)
							err = l.ParentSpeedRatio.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								s.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("ratio", l.ParentSpeedRatio.SpeedRatio).Msg("parentspeedratio")

						case server.WishlistIntervalCode:
							l.WishlistInterval = new(server.WishlistInterval)
							err = l.WishlistInterval.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								s.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("interval", l.WishlistInterval.Interval).Msg("wishlistinterval")

						case server.PrivilegedUsersCode:
							l.PrivilegedUsers = new(server.PrivilegedUsers)
							err = l.PrivilegedUsers.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								s.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("users", len(l.PrivilegedUsers.Users)).Msg("privilegedusers")

						case server.ExcludedSearchPhrasesCode:
							l.ExcludedSearchPhrases = new(server.ExcludedSearchPhrases)
							err = l.ExcludedSearchPhrases.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								s.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("phrases", len(l.ExcludedSearchPhrases.Phrases)).Msg("excludedsearchphrases")

						case server.CheckPrivilegesCode:
							l.CheckPrivileges = new(server.CheckPrivileges)
							err = l.CheckPrivileges.Deserialize(r[0])
							if err != nil && !errors.Is(err, io.EOF) {
								s.mu.Unlock()
								return
							}

							log.Debug().Any("code", code).Int("timeleft", l.CheckPrivileges.TimeLeft).Msg("checkprivileges")
						}

						s.m[code] = s.m[code][1:]
						necessaryCodes[code] = true
					}
				}
				s.mu.Unlock()
			}(code)
		}

		time.Sleep(200 * time.Millisecond)
	}
}
