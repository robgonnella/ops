# Opi
## On-Prem Server Management

## Getting Started

- Install dependencies

```bash
brew install go
brew install nmap
brew install ansible
```

- Ensure the path variables are set for golang:

```
export GOPATH="$HOME/go"
PATH="${GOPATH}/bin:$PATH"
```

- Build opi

```bash
make opi
```

- launch opi ui and scan network

```bash
./build/opi
```

- clear database file and log file

```bash
./build/opi clean
```
