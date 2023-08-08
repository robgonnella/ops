package discovery

import (
	"context"
)

//go:generate mockgen -destination=../mock/discovery/mock_discovery.go -package=mock_discovery . DetailScanner,Scanner

// DetailScanner interface for gathering more details about a device
type DetailScanner interface {
	GetServerDetails(ctx context.Context, ip string) (*Details, error)
}

// Scanner interface for scanning a network for devices
type Scanner interface {
	Scan() error
	Stop()
}

// Service interface for monitoring a network
type Service interface {
	MonitorNetwork()
	Stop()
}
