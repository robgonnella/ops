package discovery

import (
	"context"

	"github.com/robgonnella/ops/internal/server"
)

//go:generate mockgen -destination=../mock/discovery/mock_discovery.go -package=mock_discovery . DetailScanner,PacketScanner,Scanner

// DetailScanner interface for gathering more details about a device
type DetailScanner interface {
	GetServerDetails(ctx context.Context, ip string) (*Details, error)
}

// PacketScanner
type PacketScanner interface {
	ListenForPackets(resultChan chan *server.Server)
}

// Scanner interface for scanning a network for devices
type Scanner interface {
	Scan(resultChan chan *DiscoveryResult) error
	Stop()
}

// Service interface for monitoring a network
type Service interface {
	MonitorNetwork()
	Stop()
}
