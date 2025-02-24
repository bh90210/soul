# Soul

[![Go Reference](https://pkg.go.dev/badge/github.com/bh90210/soul.svg)](https://pkg.go.dev/github.com/bh90210/soul)
[![codecov](https://codecov.io/gh/bh90210/soul/graph/badge.svg?token=1VXJR0HV3C)](https://codecov.io/gh/bh90210/soul)
[![Go Report](https://goreportcard.com/badge/github.com/bh90210/soul)](https://goreportcard.com/report/github.com/bh90210/soul)

A golang implementation of the soulseek protocol.

# Protocol Specification

This implementation and naming is based on the [Nicotine+](https://nicotine-plus.github.io/nicotine-plus/doc/SLSKPROTOCOL.html) documentation but [aioslsk](https://aioslsk.readthedocs.io) was also consulted regularly too.

# How to use



## Tests

The library is thoroughly covered via unit and integration tests. Running the `-short` test suite will result in running the unit tests. The integration suite needs some preparation before it can be used.

### Unit
```bash
go test -race -count 10 -parallel 10 -short --cover -covermode=atomic -coverpkg=./... ./... -coverprofile=coverage.txt
```
### Integration

_Note that the integration tests -as a necessity- are using the real network. Run them sparsely._

Tests impersonate two users exchanging messages and files and thus we need two ports open (2234 & 2235.)

```bash
go test --cover -covermode=atomic -coverpkg=./... ./... -coverprofile=coverage.txt	
```
# Acknowledgements

Fork of https://github.com/a-cordier/goose. Thanks to a-cordier for starting the effort as this is usually the hardest part.
