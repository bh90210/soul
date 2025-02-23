package flow

import "github.com/rs/zerolog"

type Config struct {
	SoulseekAddress string
	SoulseekPort    int
	Username        string
	Password        string
	SharedFolders   int
	SharedFiles     int
	LogLevel        zerolog.Level
}
