package discovery

// ServerStatus represents possible server statuses
type ServerStatus string

const (
	// ServerOnline status used when a server online
	ServerOnline ServerStatus = "online"
	// ServerOffline status used when a server offline
	ServerOffline ServerStatus = "offline"
)

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
	Type     string
	ID       string
	Hostname string
	IP       string
	OS       string
	Vendor   string
	Status   ServerStatus
	Port     Port
}

// Details represents the details returned by DetailScanner
type Details struct {
	Hostname string
	OS       string
}
