package discovery

import "github.com/robgonnella/ops/internal/server"

// DeviceType represents a type of device discovered on the network
type DeviceType int

// Enum values for our different device types
const (
	ServerDevice DeviceType = iota
	UnknownDevice
)

type PortStatus string

const (
	PortOpen   PortStatus = "open"
	PortClosed PortStatus = "closed"
)

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
	Ports    []Port
}

type Details struct {
	Hostname string
	OS       string
}
