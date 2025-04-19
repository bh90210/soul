# Soul

[![Go Reference](https://pkg.go.dev/badge/github.com/bh90210/soul.svg)](https://pkg.go.dev/github.com/bh90210/soul)
[![codecov](https://codecov.io/gh/bh90210/soul/graph/badge.svg?token=1VXJR0HV3C)](https://codecov.io/gh/bh90210/soul)
[![Go Report](https://goreportcard.com/badge/github.com/bh90210/soul)](https://goreportcard.com/report/github.com/bh90210/soul)
![](https://github.com/bh90210/soul/actions/workflows/tests.yaml/badge.svg)

A Go implementation of the SoulSeek protocol.

# Protocol Specification

This implementation and naming convention is based on the [Nicotine+](https://nicotine-plus.github.io/nicotine-plus/doc/SLSKPROTOCOL.html) documentation but [aioslsk](https://aioslsk.readthedocs.io) was also consulted regularly too.

The library offers complete coverage of all server, peer, distributed and file messages' serialization and deserialization.

On top there is a client package with file sharing capabilities (global search/download/upload API/distributed network participation, obfuscation) but no room searches, direct peer searches for directories or wishlist (PRs are welcome* :)

_* Even though server messages facilitating chat functionality are present in the library for completeness sake, consider not using them as they are unencrypted in the open._

# How to use

SoulSeek protocol has 4 different connection types. Server, Peer, File and Distributed. For each connection type there is a unique set of messages the server and peers expect from us and vice versa.

For a simple search & download example see `/cmd/search-download`.

## Low level

Low level code facilitating the serialization and deserialization of each connection type and message code lives under the `server`, `peer`, `file` and `distributed` packages respectively. Each package offers a pair of Read/Write functions and the complete in use message codes (I did not implement obsolete protocol message codes.)

Each message is a struct, for example to make use of the server connection Login message code you need to:
```go
package main

import (
	"fmt"
	"net"

	"github.com/bh90210/soul/server"
)

func main() {
	// Open a connection with the server.
	conn, _ := net.Dial("tcp", "server.slsknet.org:2242")

	// Send the login message.
	server.Write(conn, &server.Login{Username: "username", Password: "password"})

	// Receive the server response.
	res, _, code, _ := server.Read(conn)
	switch code {
	case server.CodeLogin:
		login := new(server.Login)
		login.Deserialize(res)

		fmt.Println(login.Greet, login.IP, login.Sum)
	...
	}
}

```

## Client

To successfully make use of the network, you will need certain procedures involving multiple types of connections at once. Under `client` package you will find the most common actions a client will probably make (login, search, download, participation in the distributed network and API for responding to search requests and uploads.) If like me your goal is to make a CLI, preferably one that will run on a server rather than a desktop and used as a library inside other Go software, then client code in the `client` package can be potentially useful as is, albeit incomplete (no file indexing/management, no database for state etc, yes PRs are still very welcome!)

### Client & Peer

The methods of _Client_ and _Peer_ structs are purposefully small and simple. Both provide a `Relays` field that can produce listeners for all incoming messages. Think of them as routers for incoming messages. This can potentially be your point of departure. Use _Client_ and _Peer_ and come up with your own state solution. Except bug fixes the intention is for those structs/API to remain dormant.

### State

_State_ struct is where the "business logic" lives. Besides the public methods it provides, once connected to SoulSeek in the background it will take care of the distributed network and responding to peer and server requests.

## Tests

The library is covered via unit and integration tests. Running the `-short` test suite will result in running the unit tests. The integration tests need [Soulfind](https://github.com/soulfind-dev/soulfind) (check the `/testdata/Dockerfile.soulfind` for more) running. For convenience you can just `docker run --rm -it -p 2242:2242 ghcr.io/bh90210/soul:latest` and it will spin a Soulfind enabled container.

Units cover all connection types' serialization/deserialization and internal packages.

The `client` package contains the integration tests.

```bash
go test -parallel 100 --cover -covermode=atomic -coverpkg=./... ./... -tags=testdata
```

# Acknowledgements

Fork of [goose](https://github.com/a-cordier/goose). Thanks to `a-cordier` for starting the effort as this is usually the hardest part.

If this library was not what you were looking for consider checking out [spotseek](https://github.com/boristopalov/spotseek).

Shout-out to [slskd](https://github.com/slskd/slskd), it immensely helped with testing the implementation.

# TODO
- [ ] Finish upload.
- [ ] Finish directory/folder peer responses.
- [ ] Client integration tests (login-search-download -> login-respond to search-upload, all server messages.)
- [ ] Search code for outstanding TODOs.
- [ ] Release v1.2.0.
- [ ] Rate limits for peers: (re-)downloads, requests, connections.
- [ ] Release v1.2.1.