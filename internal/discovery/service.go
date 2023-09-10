package discovery

import (
	"context"
	"time"

	"github.com/robgonnella/go-lanscan/scanner"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/logger"
)

const (
	DiscoveryArpUpdateEvent = "DISCOVERY_ARP_UPDATE"
	DiscoverySynUpdateEvent = "DISCOVERY_SYN_UPDATE"
)

// ScannerService implements the Service interface for monitoring a network
type ScannerService struct {
	ctx           context.Context
	cancel        context.CancelFunc
	scanner       Scanner
	detailScanner DetailScanner
	resultChan    chan *scanner.ScanResult
	eventChan     chan *event.Event
	errorChan     chan error
	log           logger.Logger
}

// NewScannerService returns a new instance of ScannerService
func NewScannerService(
	scanner Scanner,
	detailScanner DetailScanner,
	resultChan chan *scanner.ScanResult,
	eventChan chan *event.Event,
) *ScannerService {
	log := logger.New()

	// Use a cancelable context so we can properly cleanup when needed
	ctxWithCancel, cancel := context.WithCancel(context.Background())

	return &ScannerService{
		ctx:           ctxWithCancel,
		cancel:        cancel,
		scanner:       scanner,
		detailScanner: detailScanner,
		resultChan:    resultChan,
		eventChan:     eventChan,
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
			switch r.Type {
			case scanner.ARPResult:
				res := r.Payload.(*scanner.ArpScanResult)
				dr := &DiscoveryResult{
					Type:     DiscoveryArpUpdateEvent,
					ID:       res.MAC.String(),
					IP:       res.IP.String(),
					Hostname: "",
					OS:       "",
					Status:   ServerOnline,
					Port: Port{
						ID:     22,
						Status: PortClosed,
					},
				}
				go s.handleDiscoveryResult(dr)
			case scanner.SYNResult:
				res := r.Payload.(*scanner.SynScanResult)
				dr := &DiscoveryResult{
					Type:     DiscoverySynUpdateEvent,
					ID:       res.MAC.String(),
					IP:       res.IP.String(),
					Hostname: "",
					OS:       "",
					Status:   ServerStatus(res.Status),
					Port: Port{
						ID:     res.Port.ID,
						Status: PortStatus(res.Port.Status),
					},
				}
				go s.handleDiscoveryResult(dr)
			}
		case err := <-s.errorChan:
			s.log.Error().Err(err).Msg("discovery service encountered an error")
			return
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
		"type":      result.Type,
		"id":        result.ID,
		"hostname":  result.Hostname,
		"ip":        result.IP,
		"os":        result.OS,
		"status":    result.Status,
		"sshStatus": result.Port.Status,
	}

	s.log.Info().Fields(fields).Msg("found network device")

	if result.Port.ID == 22 && result.Port.Status == PortOpen {
		s.log.Info().Str("ip", result.IP).Msg("retrieving device details")

		details, err := s.detailScanner.GetServerDetails(s.ctx, result.IP)

		if err == nil {
			result.Hostname = details.Hostname
			result.OS = details.OS
		} else {
			s.log.
				Error().Err(err).
				Str("ip", result.IP).
				Msg("details scan failed")
		}
	}

	if result.Hostname == "" {
		result.Hostname = "unknown"
	}

	if result.OS == "" {
		result.OS = "unknown"
	}

	go func() {
		s.eventChan <- &event.Event{
			Type:    result.Type,
			Payload: result,
		}
	}()
}
