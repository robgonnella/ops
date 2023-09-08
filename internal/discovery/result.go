package discovery

import "github.com/robgonnella/ops/internal/server"

// PortStatus represents possible port statuses
type PortStatus string

const (
	// PortOpen status used when a port is marked open
	PortOpen PortStatus = "open"
	// PortClosed status used when a port is marked closed
	PortClosed PortStatus = "closed"
)

// Port data structure representing a server port
type Port struct {
	ID     uint16
	Status PortStatus
}

// DiscoveryResult represents our discovered device on the network
type DiscoveryResult struct {
	ID       string
	Hostname string
	IP       string
	OS       string
	Status   server.Status
	Port     Port
}

// Details represents the details returned by DetailScanner
type Details struct {
	Hostname string
	OS       string
}
