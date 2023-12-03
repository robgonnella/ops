package discovery

import (
	"context"

	"github.com/robgonnella/go-lanscan/pkg/scanner"
	"github.com/robgonnella/ops/internal/config"
)

//go:generate mockgen -destination=../mock/discovery/mock_discovery.go -package=mock_discovery . DetailScanner,Scanner

// DetailScanner interface for gathering more details about a device
type DetailScanner interface {
	GetServerDetails(ctx context.Context, ip, sshPort string) (*Details, error)
}

// Scanner interface for scanning a network for devices
type Scanner interface {
	Scan() error
	Stop()
	Results() chan *scanner.ScanResult
}

// Service interface for monitoring a network
type Service interface {
	MonitorNetwork() error
	SetConfigAndScanner(conf config.Config, netScanner Scanner)
	Stop()
}
