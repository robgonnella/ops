package server

import "github.com/robgonnella/ops/internal/event"

//go:generate mockgen -destination=../mock/server/mock_server.go -package=mock_server . Repo,Service

// Status represents possible server statues
type Status string

// SSHStatus represents possible server ssh statuses
type SSHStatus string

const (
	// StatusUnknown unknown status for server
	StatusUnknown Status = "unknown"
	// StatusOnline status if server is online
	StatusOnline Status = "online"
	// StatusOffline status if server is offline
	StatusOffline Status = "offline"
	// SSHEnabled status when server has ssh enabled
	SSHEnabled SSHStatus = "enabled"
	// SSHDisabled status when server has ssh disabled
	SSHDisabled SSHStatus = "disabled"
)

// Server database model for a server
type Server struct {
	ID        string `gorm:"primaryKey"`
	Status    Status
	Hostname  string
	IP        string
	OS        string
	SshStatus SSHStatus
}

// Repo interface for accessing stored servers
type Repo interface {
	GetAllServers() ([]*Server, error)
	GetServerByID(serverID string) (*Server, error)
	GetServerByIP(ip string) (*Server, error)
	AddServer(server *Server) (*Server, error)
	UpdateServer(server *Server) (*Server, error)
	RemoveServer(id string) error
}

// Service interface for server related logic
type Service interface {
	GetAllServers() ([]*Server, error)
	GetAllServersInNetworkTargets(targets []string) ([]*Server, error)
	AddOrUpdateServer(req *Server) error
	MarkServerOffline(ip string) error
	StreamEvents(send chan *event.Event) int
	StopStream(id int)
	GetServer(id string) (*Server, error)
}
