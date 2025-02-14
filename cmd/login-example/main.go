package main

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/bh90210/soul/login"
)

func main() {
	conn, err := net.Dial("tcp", "server.slsknet.org:2242")
	if err != nil {
		log.Fatal("Unable to connect to server")
	}

	w := bufio.NewWriter(conn)
	loginMessage := login.Write("username", "password")
	w.Write(loginMessage)
	w.Flush()
	if err != nil {
		log.Fatal("login error")
	}
	res := bufio.NewReader(conn)
	response := login.Read(res)
	if response.OK() {
		fmt.Println(response.Greet, response.IP, response.Sum)
	} else {
		fmt.Println(response.Reason)
	}
}
