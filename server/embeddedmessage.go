package server

import "github.com/bh90210/soul"

const EmbeddedMessageCode soul.CodeServer = 93

type EmbeddedMessage struct {
	Code    int
	Message []byte
}

// TODO: Implement Deserialize.
