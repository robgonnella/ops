package discovery

import (
	"context"
	"time"

	"github.com/robgonnella/go-lanscan/scanner"
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
	resultChan    chan *scanner.SynScanResult
	scanDone      chan bool
	errorChan     chan error
	log           logger.Logger
}

// NewScannerService returns a new instance of ScannerService
func NewScannerService(
	scanner Scanner,
	detailScanner DetailScanner,
	serverService server.Service,
	resultChan chan *scanner.SynScanResult,
	doneChan chan bool,
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
		scanDone:      doneChan,
		errorChan:     make(chan error),
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
	s.cancel()
	s.scanner.Stop()
}

// private
// make polling calls to scanner.Scan()
func (s *ScannerService) pollNetwork() {
	ticker := time.NewTicker(time.Second * 30)

	// start first scan
	// always scan in goroutine to prevent blocking result channel
	go func() {
		if err := s.scanner.Scan(); err != nil {
			s.errorChan <- err
		}
	}()

	for {
		select {
		case <-s.ctx.Done():
			s.log.Info().Msg("Network polling stopped")
			ticker.Stop()
			return
		case r := <-s.resultChan:
			dr := &DiscoveryResult{
				ID:       r.MAC.String(),
				IP:       r.IP.String(),
				Hostname: "",
				OS:       "",
				Status:   server.Status(r.Status),
				Port: Port{
					ID:     r.Port.ID,
					Status: PortStatus(r.Port.Status),
				},
			}
			s.handleDiscoveryResult(dr)
		case err := <-s.errorChan:
			s.log.Error().Err(err).Msg("discovery service encountered an error")
			return
		case <-s.scanDone:
			// nothing to do here since we're polling
		case <-ticker.C:
			// always scan in goroutine to prevent blocking result channel
			go func() {
				if err := s.scanner.Scan(); err != nil {
					s.errorChan <- err
				}
			}()
		}
	}
}

// handle results found during polling
func (s *ScannerService) handleDiscoveryResult(result *DiscoveryResult) {
	fields := map[string]interface{}{
		"id":        result.ID,
		"hostname":  result.Hostname,
		"ip":        result.IP,
		"os":        result.OS,
		"status":    result.Status,
		"sshStatus": s.getSSHStatus(result),
	}

	s.log.Info().Fields(fields).Msg("found network device")

	switch result.Status {
	case server.StatusOnline:
		go s.setServerToOnline(result)
	case server.StatusOffline:
		go s.setServerToOffline(result)
	default:
		s.log.Info().
			Str("status", string(result.Status)).
			Msg("Device detected with unimplemented status action")
	}

}

// checks if port is 22 and whether the port is open or closed
func (s *ScannerService) getSSHStatus(result *DiscoveryResult) server.SSHStatus {
	if result.Port.ID == 22 {
		if result.Port.Status == PortOpen {
			return server.SSHEnabled
		}

		return server.SSHDisabled
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
