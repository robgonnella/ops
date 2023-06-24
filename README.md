![](assets/logo-no-background-small.png)

## On-Prem Server Management

Ops is a terminal UI application for managing on-premise bare-metal servers
and VMs. It allows you see servers currently on your network and quickly ssh to
them. Within the UI, you can create and manage multiple network configurations,
you can choose a default set of ssh credentials to use for all servers, and you
can override those defaults for individual IPs where needed.

This project is heavily inspired by [derailed]'s amazing work on [k9s] for
managing kubernetes clusters via a terminal ui application.

## Installation

- homebrew

```bash
brew install ops
```

- source build & pre-built binaries
  - dependencies
    - golang
    - make
    - ansible
    - nmap
  - install pre-built binaries
    - https://github.com/robgonnella/ops/releases
  - build from source
```bash
git clone https://github.com/robgonnella/ops.git
cd ops
make install
```

## Usage

On first launch a default configuration will be generated based on your machines
default network settings. If your machine is not connected to a network the app
will fail to start.

- start application

```bash
ops
```

- clear database file and log file

```bash
ops clear
```

## Files and Config

- `ops.db`: Configurations and server details are stored locally on your machine
  in a sqlite database located in your machines default cache directory. On Unix
  systems, it returns `$XDG_CACHE_HOME` if non-empty, else `$HOME/.cache`. On
  Darwin, it returns `$HOME/Library/Caches`.
- `ops.log`: Additional logging `~/.config/ops/ops.log`

## Technologies

- [tview] is used to build the frontend. This is a wonderful open source
  terminal ui library provided by [rivo]!
- [nmap] is used on the backend to perform arp scanning of networks to find
  and track devices.
- [ansible] is also used on the backend to gather additional details about a
  device where ssh is granted

[rivo]: https://github.com/rivo
[tview]: https://github.com/rivo/tview
[ansible]: https://docs.ansible.com/
[nmap]: https://nmap.org/
[k9s]: https://github.com/derailed/k9s
[derailed]: https://github.com/derailed
