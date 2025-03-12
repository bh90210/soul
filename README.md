# Soul

[![Go Reference](https://pkg.go.dev/badge/github.com/bh90210/soul.svg)](https://pkg.go.dev/github.com/bh90210/soul)
[![codecov](https://codecov.io/gh/bh90210/soul/graph/badge.svg?token=1VXJR0HV3C)](https://codecov.io/gh/bh90210/soul)
[![Go Report](https://goreportcard.com/badge/github.com/bh90210/soul)](https://goreportcard.com/report/github.com/bh90210/soul)
![](https://github.com/bh90210/soul/actions/workflows/tests.yaml/badge.svg)

A Go implementation of the SoulSeek protocol.

# Protocol Specification

This implementation and naming convention is based on the [Nicotine+](https://nicotine-plus.github.io/nicotine-plus/doc/SLSKPROTOCOL.html) documentation but [aioslsk](https://aioslsk.readthedocs.io) was also consulted regularly too.

The library offers complete coverage of all server, peer, distributed and file messages' serialization and deserialization.

On top there is a client package with file sharing capabilities (global search/download/upload API/distributed network participation, obfuscation) but no chat, rooms, room searches & direct messages/searches (PRs are welcome :).

# How to use

SoulSeek protocol has 4 different connection types. Server, Peer, File and Distributed. For each connection type there is a unique set of messages the server and peers expect from us and vice versa.

For a simple search & download example see `/cmd/search-download`.

## Low level

Low level code facilitating the serialization and deserialization of each connection type and message code lives under the `server`, `peer`, `file` and `distributed` packages respectively. Each package offers a pair of Read/Write functions and the complete in use message codes (I did not implement obsolete protocol message codes.)

Each message is a struct, for example to make use of the server connection Login message code you need to:
```go
// Open a connection with the server.
conn, _ = net.Dial("tcp", "server.slsknet.org:2242")

// Send the login message.
login := new(server.Login)
req, _ := login.Serialize("username", "password")
server.MessageWrite(conn, req)

// Receive the server message.
res, _, code, _ := server.MessageRead(conn)
switch server.Code(code) { // Uint32 to own server.Code type (int.)
case server.CodeLogin:
	login := new(server.Login)
	login.Deserialize(res)
	fmt.Println(login.Greet, login.IP, login.Sum)
...
}
```

## Client

To successfully make use of the network, you will need certain procedures involving multiple types of connections at once. Under `client` package you will find the most common actions a client will probably make (login, search, download, participation in the distributed network and API for responding to search quests and uploads.) If like me your goal is to make a CLI, preferably one that will run on a server rather than a desktop and used as a library inside other Go software, then client code in the `client` package can be potentially useful as is, albeit incomplete (no private messages, not chat rooms, no file indexing/handling, no database etc, and yes PRs are still very welcome!)

### API

#### Client

#### Peer

#### State

## Tests

The library is thoroughly covered via unit and integration tests. Running the `-short` test suite will result in running the unit tests. The integration tests need docker available in host to run. This is because we are opening a local [Soulfind](https://github.com/soulfind-dev/soulfind) instance (check the `/client/Dockerfile` for more.)

Units cover all connection types' serialization/deserialization and internal packages.

The `client` package contains the integration tests.

```bash
go test -parallel 100 --cover -covermode=atomic -coverpkg=./... ./...
```

# Acknowledgements

Fork of [goose](https://github.com/a-cordier/goose). Thanks to `a-cordier` for starting the effort as this is usually the hardest part.

If this library was not what you were looking for consider checking out [spotseek](https://github.com/boristopalov/spotseek).