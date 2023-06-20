package server

import "github.com/robgonnella/ops/internal/event"

//go:generate mockgen -destination=../mock/server/mock_server.go -package=mock_server . Repo,Service

type Status string
type SSHStatus string

const (
	StatusUnknown Status    = "unknown"
	StatusOnline  Status    = "online"
	StatusOffline Status    = "offline"
	SSHEnabled    SSHStatus = "enabled"
	SSHDisabled   SSHStatus = "disabled"
)

type Server struct {
	ID        string
	Status    Status
	Hostname  string
	IP        string
	OS        string
	SshStatus SSHStatus
}

type Repo interface {
	GetAllServers() ([]*Server, error)
	GetServerByID(serverID string) (*Server, error)
	GetServerByIP(ip string) (*Server, error)
	AddServer(server *Server) (*Server, error)
	UpdateServer(server *Server) (*Server, error)
	RemoveServer(id string) error
}

type Service interface {
	GetAllServers() ([]*Server, error)
	AddOrUpdateServer(req *Server) error
	MarkServerOffline(ip string) error
	StreamEvents(send chan *event.Event) int
	StopStream(id int)
	GetServer(id string) (*Server, error)
}
