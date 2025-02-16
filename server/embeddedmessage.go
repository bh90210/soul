package server

const EmbeddedMessageCode Code = 93

type EmbeddedMessage struct {
	Code    int
	Message []byte
}

// TODO: Implement Deserialize.
