# Soul

[![Go Reference](https://pkg.go.dev/badge/github.com/bh90210/soul.svg)](https://pkg.go.dev/github.com/bh90210/soul)
[![codecov](https://codecov.io/gh/bh90210/soul/graph/badge.svg?token=1VXJR0HV3C)](https://codecov.io/gh/bh90210/soul)
[![Go Report](https://goreportcard.com/badge/github.com/bh90210/soul)](https://goreportcard.com/report/github.com/bh90210/soul)

A Go implementation of the SoulSeek protocol.

# Protocol Specification

This implementation and naming convention is based on the [Nicotine+](https://nicotine-plus.github.io/nicotine-plus/doc/SLSKPROTOCOL.html) documentation but [aioslsk](https://aioslsk.readthedocs.io) was also consulted regularly too.

# How to use

SoulSeek protocol has 4 different connection types. Server, Peer, File and Distributed. For each connection type there is a unique set of messages other clients expect from us and vice versa.

## Low level

Low level code facilitating the serialization and deserialization of each connection type and message code lives under the `server`, `peer`, `file` and `distributed` packages respectively. Each package offers a pair of Read/Write functions and the complete in use message codes (I did not implement obsolete protocol message codes.)

Each message is a struct, for example to make use of the server connection Login message code you need to:
```go
login := new(server.Login)
message, _ := login.Serialize("username", "password")
server.MessageWrite(tcpConn, message)
```

And to receive the server response:
```go
reader, _, messageCode, _ := server.MessageRead(tcpConn)
switch messageCode {
case server.LoginCode:
	login := new(server.Login)
	login.Deserialize(reader)
	fmt.Println(login.Greet, login.IP, login.Sum)
}
```

## Flow

To successfully make use of the network certain procedures involving multiple types of connections at once are needed. Under `flow` package you will find the most common actions you will probably make (login, search, download, upload.) If like me your goal is to make a CLI, preferably one that will run on a server not desktop, then code in the `flow` package can be potentially useful as is, albeit incomplete (no private messages, not chat rooms etc.) 

If your goal is to make a full GUI desktop client then code under `flow` at best can server as a guide but you will probably want to write your own message handling system (a message queue of some sorts seems like an obvious solution for example.)

## Tests

The library is thoroughly covered via unit and integration tests. Running the `-short` test suite will result in running the unit tests. The integration suite needs some preparation before it can be used.

Unit suite covers all connection types' serialization/deserialization functions and internal packages.

The `flow` package contains the integration test suite.

### Unit

Unit tests can make use of the `-race` and `-parallel` flags.

```bash
go test -short -race -count 10 -parallel 10 --cover -covermode=atomic -coverpkg=./... ./... -coverprofile=coverage.txt
```

### Integration

_Note that the integration tests -as a necessity- are using the real network. Run them sparsely._

Tests impersonate two users exchanging messages and files and thus we need two ports open (2234 & 2235.) As such integration tests can not make use of `-race` and `-parallel`.

```bash
go test --cover -covermode=atomic -coverpkg=./... ./... -coverprofile=coverage.txt	
```

# Acknowledgements

Fork of https://github.com/a-cordier/goose. Thanks to a-cordier for starting the effort as this is usually the hardest part.
