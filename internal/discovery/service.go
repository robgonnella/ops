package discovery

import (
	"context"
	"time"

	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/server"
)

// ScannerService implements our discovery service using nmap
type ScannerService struct {
	ctx           context.Context
	cancel        context.CancelFunc
	scanner       Scanner
	detailScanner DetailScanner
	serverService server.Service
	log           logger.Logger
}

// NewScannerService returns a new intance of nmap network discovery ScannerService
func NewScannerService(scanner Scanner, detailScanner DetailScanner, serverService server.Service) *ScannerService {
	log := logger.New()

	// Use a cancelable context so we can properly cleanup when needed
	ctxWithCancel, cancel := context.WithCancel(context.Background())

	return &ScannerService{
		ctx:           ctxWithCancel,
		cancel:        cancel,
		scanner:       scanner,
		detailScanner: detailScanner,
		serverService: serverService,
		log:           log,
	}
}

// MonitorNetwork polls the network and calls out to grpc with the results
func (s *ScannerService) MonitorNetwork() {
	s.log.Info().Msg("Starting network discovery")

	// blocking call that continuously scans the network on an interval
	s.pollNetwork()
}

// Stop stop network discover
func (s *ScannerService) Stop() {
	s.scanner.Stop()
	s.cancel()
}

// private
// pollNetwork runs Discover function on an interval to discover devices on the network
func (s *ScannerService) pollNetwork() {
	pollTime := time.Second * 30

	for {
		select {
		case <-s.ctx.Done():
			s.log.Info().Msg("Network polling stopped")
			s.cancel()
			return
		default:
			results, err := s.scanner.Scan()

			if err != nil {
				s.log.Warn().Err(err).Msg("Error polling network")
			} else {
				s.log.Info().
					Fields(map[string]interface{}{"count": len(results)}).
					Msg("Discovery results")

				s.handleDiscoveryResults(results)
			}

			time.Sleep(pollTime)
		}
	}
}

func (s *ScannerService) handleDiscoveryResults(results []*DiscoveryResult) {
	for _, result := range results {
		deviceType := s.getDeviceType(result)

		fields := map[string]interface{}{
			"id":         result.ID,
			"hostname":   result.Hostname,
			"ip":         result.IP,
			"os":         result.OS,
			"status":     result.Status,
			"sshStatus":  s.getSSHStatus(result),
			"deviceType": deviceType,
		}

		s.log.Info().Fields(fields).Msg("found network device")

		switch result.Status {
		case server.StatusOnline:
			if deviceType == ServerDevice {
				go s.setServerToOnline(result)
			} else {
				s.log.Info().
					Str("ip", result.IP).
					Msg("Unknown device detected on network")
			}
		case server.StatusOffline:
			go s.setServerToOffline(result)
		default:
			s.log.Info().
				Str("status", string(result.Status)).
				Msg("Device detected with unimplemented status action")
		}
	}
}

func (s *ScannerService) getDeviceType(result *DiscoveryResult) DeviceType {
	for _, port := range result.Ports {
		if port.ID == 22 {
			return ServerDevice
		}
	}

	return UnknownDevice
}

func (s *ScannerService) getSSHStatus(result *DiscoveryResult) server.SSHStatus {
	for _, port := range result.Ports {
		if port.ID == 22 {
			if port.Status == PortOpen {
				return server.SSHEnabled
			}

			return server.SSHDisabled
		}
	}

	return server.SSHDisabled
}

// makes a call to our grpc server with a request to update the servers status
// to online. We also checke that server's associated services and update each
// service's status accordingly.
func (s *ScannerService) setServerToOnline(result *DiscoveryResult) {
	// finally update our server to "online"
	sshStatus := s.getSSHStatus(result)

	if sshStatus == server.SSHEnabled {
		s.log.Info().Str("ip", result.IP).Msg("retrieving device details")

		details, err := s.detailScanner.GetServerDetails(s.ctx, result.IP)

		if err == nil {
			result.Hostname = details.Hostname
			result.OS = details.OS
		} else {
			s.log.
				Error().Err(err).
				Str("ip", result.IP).
				Msg("ansible scan failed")
		}
	}

	if result.Hostname == "" {
		result.Hostname = "unknown"
	}

	if result.OS == "" {
		result.OS = "unknown"
	}

	s.log.Info().
		Str("ip", result.IP).
		Msg("marking device online")

	err := s.serverService.AddOrUpdateServer(&server.Server{
		ID:        result.ID,
		Hostname:  result.Hostname,
		IP:        result.IP,
		OS:        result.OS,
		Status:    result.Status,
		SshStatus: sshStatus,
	})

	if err != nil {
		s.log.Error().Err(err).Msg("error marking device online")
	}
}

func (s *ScannerService) setServerToOffline(result *DiscoveryResult) {
	s.log.Info().Str("ip", result.IP).Msg("Marking device offline")
	if err := s.serverService.MarkServerOffline(result.IP); err != nil {
		s.log.Error().Err(err).Msg("error marking device offline")
	}
}
