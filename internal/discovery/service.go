package discovery

import (
	"context"
	"time"

	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/server"
)

// ScannerService implements the Service interface for monitoring a network
type ScannerService struct {
	ctx           context.Context
	cancel        context.CancelFunc
	scanner       Scanner
	detailScanner DetailScanner
	serverService server.Service
	resultChan    chan *DiscoveryResult
	log           logger.Logger
}

// NewScannerService returns a new instance of ScannerService
func NewScannerService(
	scanner Scanner,
	detailScanner DetailScanner,
	serverService server.Service,
	resultChan chan *DiscoveryResult,
) *ScannerService {
	log := logger.New()

	// Use a cancelable context so we can properly cleanup when needed
	ctxWithCancel, cancel := context.WithCancel(context.Background())

	return &ScannerService{
		ctx:           ctxWithCancel,
		cancel:        cancel,
		scanner:       scanner,
		detailScanner: detailScanner,
		serverService: serverService,
		resultChan:    resultChan,
		log:           log,
	}
}

// MonitorNetwork polls the network to discover and track devices
func (s *ScannerService) MonitorNetwork() {
	s.log.Info().Msg("Starting network discovery")

	// blocking call that continuously scans the network on an interval
	s.pollNetwork()
}

// Stop stop network discover. Once called this service will be useless.
// A new one must be instantiated to continue
func (s *ScannerService) Stop() {
	s.scanner.Stop()
	s.cancel()
}

// private
// make polling calls to scanner.Scan()
func (s *ScannerService) pollNetwork() {
	ticker := time.NewTicker(time.Second * 30)

	// start first scan
	// always scan in goroutine to prevent blocking result channel
	go func() {
		if err := s.scanner.Scan(); err != nil {
			s.cancel()
		}
	}()

	for {
		select {
		case <-s.ctx.Done():
			s.log.Info().Msg("Network polling stopped")
			ticker.Stop()
			s.cancel()
			return
		case r := <-s.resultChan:
			s.handleDiscoveryResult(r)
		case <-ticker.C:
			// always scan in goroutine to prevent blocking result channel
			go func() {
				if err := s.scanner.Scan(); err != nil {
					s.cancel()
				}
			}()
		}
	}
}

// handle results found during polling
func (s *ScannerService) handleDiscoveryResult(result *DiscoveryResult) {
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

// if port 22 is detected then we can assume its a server
func (s *ScannerService) getDeviceType(result *DiscoveryResult) DeviceType {
	for _, port := range result.Ports {
		if port.ID == 22 {
			return ServerDevice
		}
	}

	return UnknownDevice
}

// checks if port is 22 and whether the port is open or closed
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

// makes a call to our server service to update the servers status to online
func (s *ScannerService) setServerToOnline(result *DiscoveryResult) {
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

// makes a call to our server service to set a device to "offline"
func (s *ScannerService) setServerToOffline(result *DiscoveryResult) {
	s.log.Info().Str("ip", result.IP).Msg("Marking device offline")
	if err := s.serverService.MarkServerOffline(result.IP); err != nil {
		s.log.Error().Err(err).Msg("error marking device offline")
	}
}
