package discovery

import "context"

//go:generate mockgen -destination=../mock/discovery/mock_discovery.go -package=mock_discovery . DetailScanner,Scanner

type DetailScanner interface {
	GetServerDetails(ctx context.Context, ip string) (*Details, error)
}

type Scanner interface {
	Scan() ([]*DiscoveryResult, error)
	Stop()
}

type Service interface {
	MonitorNetwork()
	Stop()
}
