package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/bh90210/soul/server"
)

func main() {
	conn, err := net.Dial("tcp", "server.slsknet.org:2242")
	// conn, err := net.Dial("tcp", "localhost:2242") // Local dev.
	if err != nil {
		log.Fatal("server connection")
	}

	go func() {
		for {
			r, _, code, err := server.ReadMessage(conn)
			if err != nil {
				log.Fatal("readmessage", err)
			}

			switch code {
			case server.LoginCode:
				login := new(server.Login)
				err := login.Deserialize(r)
				if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
					log.Fatal("loginerr ", err)
				}

				if err == nil {
					fmt.Println("login: ", login.Greet, login.IP, login.Sum)
				}

			case server.GetPeerAddressCode:
				peerAddress := new(server.GetPeerAddress)
				err := peerAddress.Deserialize(r)
				// TODO: fix the unexpected EOF error.
				if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
					log.Fatal("peeraddresserr ", err)
				}

				fmt.Println("username: ", peerAddress.Username, peerAddress.IP, peerAddress.Port, peerAddress.ObfuscatedPort)

			case server.RoomListCode:
				roomList := new(server.RoomList)
				err := roomList.Deserialize(r)
				if err != nil && !errors.Is(err, io.EOF) {
					log.Fatal("roomlisterr", err)
				}

				for _, v := range roomList.Rooms {
					fmt.Println("room", v.Name, v.Users, v.Private, v.Owned, v.Operated)
				}

			case server.PrivilegedUsersCode:
				privilegedUsers := new(server.PrivilegedUsers)
				err := privilegedUsers.Deserialize(r)
				if err != nil && !errors.Is(err, io.EOF) {
					log.Fatal("privilegeduserserr", err)
				}

				fmt.Println("privileged", len(privilegedUsers.Users))

			case server.ParentMinSpeedCode:
				parentMinSpeedResponse := new(server.ParentMinSpeed)
				err := parentMinSpeedResponse.Deserialize(r)
				if err != nil && !errors.Is(err, io.EOF) {
					log.Fatal("parentminspeederr", err)
				}

				fmt.Println("minspeed", parentMinSpeedResponse.MinSpeed)

			case server.ParentSpeedRatioCode:
				parentSpeedRatio := new(server.ParentSpeedRatio)
				err := parentSpeedRatio.Deserialize(r)
				if err != nil && !errors.Is(err, io.EOF) {
					log.Fatal("parentspeedratioerr", err)
				}

				fmt.Println("speedratio", parentSpeedRatio.SpeedRatio)

			case server.WishlistIntervalCode:
				wishlistInterval := new(server.WishlistInterval)
				err := wishlistInterval.Deserialize(r)
				if err != nil && !errors.Is(err, io.EOF) {
					log.Fatal("wishlistintervalerr", err)
				}

				fmt.Println("wishlistinterval", wishlistInterval.Interval)

			case server.ExcludedSearchPhrasesCode:
				excludedSearchPhrases := new(server.ExcludedSearchPhrases)
				err := excludedSearchPhrases.Deserialize(r)
				if err != nil && !errors.Is(err, io.EOF) {
					log.Fatal("excludedsearchphraseserr", err)
				}

				for _, v := range excludedSearchPhrases.Phrases {
					fmt.Println("phrase", v)
				}

			default:
				fmt.Println("Unknown code", code)
			}
		}
	}()

	// Login.
	fmt.Println("sending login")

	login := new(server.Login)
	loginMessage, err := login.Serialize("username", "password")
	if err != nil {
		log.Fatal("login serialize")
	}

	_, err = server.SendMessage(conn, loginMessage)
	if err != nil {
		log.Fatal("write1")
	}

	// Port.
	fmt.Println("sending port")

	port := new(server.SetListenPort)
	portMessage, err := port.Serialize(2234)
	if err != nil {
		log.Fatal("port serialize")
	}

	_, err = server.SendMessage(conn, portMessage)
	if err != nil {
		log.Fatal("write2")
	}

	time.Sleep(2 * time.Second)

	// Peer's address.
	fmt.Println("requesting peer address")

	peerAddress := new(server.GetPeerAddress)
	getPeerAddressMessage, err := peerAddress.Serialize("username")
	if err != nil {
		log.Fatal("peeraddress serialize")
	}

	_, err = server.SendMessage(conn, getPeerAddressMessage)
	if err != nil {
		log.Fatal("write3")
	}

	time.Sleep(10 * time.Second)
}
