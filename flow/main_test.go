package flow

import (
	"flag"
	"log"
	"os"
	"testing"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	user1 *Client
	user2 *Client
)

func TestMain(m *testing.M) {
	flag.Parse()

	if testing.Short() {
		os.Exit(0)
	}

	username, _ := gonanoid.Generate("soulseek", 8)
	password, _ := gonanoid.Generate("124356809absvcuirpkf", 8)

	user1 = new(Client)
	user1.Config = &Config{
		SoulseekAddress: "server.slsknet.org",
		SoulseekPort:    2242,
		OwnPort:         2234,
		Username:        username,
		Password:        password,
		SharedFolders:   1,
		SharedFiles:     10,
	}

	err := user1.Dial()
	if err != nil {
		log.Fatal(err)
	}

	defer user1.Close()

	go func() {
		for {
			user1.NextMessage()
		}
	}()

	// username, _ = gonanoid.Generate("soulseek", 8)
	// password, _ = gonanoid.Generate("124356809absvcuirpkf", 8)

	// user2 = new(Client)
	// user2.Config = &Config{
	// 	SoulseekAddress: "server.slsknet.org",
	// 	SoulseekPort:    2242,
	// 	OwnPort:         2235,
	// 	Username:        username,
	// 	Password:        password,
	// 	SharedFolders:   1,
	// 	SharedFiles:     10,
	// }

	// err = user2.Dial()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer user2.Close()

	// go func() {
	// 	for {
	// 		user2.NextMessage()
	// 	}
	// }()

	m.Run()
}
