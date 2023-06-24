![](assets/logo-no-background-small.png)

# Contributing

## Prerequisites

- install dependencies

```bash
brew install make
brew install go
brew install nmap
brew install ansible
```

- Ensure the path variables are set for golang:

```bash
export GOPATH="$HOME/go"
PATH="${GOPATH}/bin:$PATH"
```

- Install golang dependencies

```bash
go install github.com/golang/mock/mockgen@v1.6.0
```

## Test

Mocks are generated using the `//go:generate` directive. Any
file that contains an interface that you would like to mock
should contain a go generate comment directive at the top of
the file. Mocks will then be generated via `make mock`

i.e.

```go
// config/interface.go
package config

//go:generate mockgen -destination=../mock/config/mock_config.go -package=mock_config . Repo,Service
```

- Generate mocks

```bash
make mock
```

- Run tests

```bash
make test
```

## Build

```bash
make ops
```

## Run

```bash
./build/ops
```

- clear database file and log file

```bash
./build/ops clean
```
