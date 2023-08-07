![](assets/logo-no-background-small.png)

## On-Prem Server Management

Ops is a terminal UI application for managing on-premise bare-metal servers
and VMs. It allows you see servers currently on your network and quickly ssh to
them. Within the UI, you can create and manage multiple network configurations,
you can choose a default set of ssh credentials to use for all servers, and you
can override those defaults for individual IPs where needed.

This project is heavily inspired by [derailed]'s amazing work on [k9s] for
managing kubernetes clusters via a terminal ui application.

## Runtime Dependencies

Ops has one external runtime dependencies, [ansible].

- mac
```bash
brew install ansible
```

- debian
```bash
sudo apt update && sudo apt install -y ansible
```

## Build Dependencies

If building Ops from source you will need a few other build dependencies.

- mac
```bash
brew install make go git
```

- debian
```bash
sudo apt update && sudo apt install -y make golang git
```

## Installation

When installing using golang or building from source, you may want to add the
following lines to one of your rc files to add your user's go bin to the
PATH variable. This will make the `ops` command available in your shell.

```bash
export GOPATH="$HOME/go"
PATH="${GOPATH}/bin:$PATH"
```

- install using golang
  - dependencies
    - golang
    - ansible
    - git

```bash
go install github.com/robgonnella/ops@latest
```

- build from source
  - dependencies
    - golang
    - make
    - ansible
    - git

```bash
git clone https://github.com/robgonnella/ops.git
cd ops
make install
```

- use pre-built binaries: https://github.com/robgonnella/ops/releases
  - dependencies
    - ansible

## Usage

On first launch a default configuration will be generated based on your machines
default network settings. If your machine is not connected to a network the app
will fail to start.

- start application

```bash
ops
```

- print version and other details

```bash
ops version
ops info
```

- clear database file and log file

```bash
ops clear
```

- show help / usage

```bash
ops help
# or
ops <cmd> --help
```

## Demo

![](assets/ops-demo.gif)

## Files and Config

- `ops.db`: Configurations and server details are stored locally on your machine
  in a sqlite database located in your machines default cache directory. On Unix
  systems, it returns `$XDG_CACHE_HOME` if non-empty, else `$HOME/.cache`. On
  Darwin, it returns `$HOME/Library/Caches`.
- `ops.log`: Additional logging `~/.config/ops/ops.log`

## Technologies

- [tview] is used to build the frontend. This is a wonderful open source
  terminal ui library provided by [rivo]!
- [ansible] is also used on the backend to gather additional details about a
  device where ssh is granted

[rivo]: https://github.com/rivo
[tview]: https://github.com/rivo/tview
[ansible]: https://docs.ansible.com/
[k9s]: https://github.com/derailed/k9s
[derailed]: https://github.com/derailed
