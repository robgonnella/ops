![](assets/logo-no-background.png)

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

- Build ops

```bash
make ops
```

- launch ops ui and scan network

```bash
./build/ops
```

- clear database file and log file

```bash
./build/ops clear
```
