package discovery

import (
	"context"
	"strconv"
	"time"

	"github.com/robgonnella/go-lanscan/pkg/scanner"
	"github.com/robgonnella/ops/internal/config"
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
	conf          config.Config
	scanner       Scanner
	detailScanner DetailScanner
	resultChan    chan *scanner.ScanResult
	pauseChan     chan struct{}
	eventManager  event.Manager
	errorChan     chan error
	monitoring    bool
	log           logger.Logger
}

// NewScannerService returns a new instance of ScannerService
func NewScannerService(
	conf config.Config,
	scanner Scanner,
	detailScanner DetailScanner,
	resultChan chan *scanner.ScanResult,
	eventManager event.Manager,
) *ScannerService {
	log := logger.New()

	// Use a cancelable context so we can properly cleanup when needed
	ctxWithCancel, cancel := context.WithCancel(context.Background())

	return &ScannerService{
		ctx:           ctxWithCancel,
		cancel:        cancel,
		conf:          conf,
		scanner:       scanner,
		detailScanner: detailScanner,
		resultChan:    resultChan,
		eventManager:  eventManager,
		errorChan:     make(chan error),
		pauseChan:     make(chan struct{}),
		monitoring:    false,
		log:           log,
	}
}

// MonitorNetwork polls the network to discover and track devices
func (s *ScannerService) MonitorNetwork() error {
	s.log.Info().Msg("Starting network discovery")

	// blocking call that continuously scans the network on an interval
	return s.pollNetwork()
}

// Stop stop network discover. Once called this service will be useless.
// A new one must be instantiated to continue
func (s *ScannerService) Stop() {
	s.cancel()
	s.scanner.Stop()
}

func (s *ScannerService) SetConfig(conf config.Config) {
	if s.monitoring {
		s.pause()
		defer func() {
			go s.pollNetwork()
		}()
	}
	s.conf = conf
}

// private
// make polling calls to scanner.Scan()
func (s *ScannerService) pollNetwork() error {
	ticker := time.NewTicker(time.Second * 30)

	defer func() {
		ticker.Stop()
		s.monitoring = false
	}()

	// start first scan
	// always scan in goroutine to prevent blocking result channel
	go func() {
		if err := s.scanner.Scan(); err != nil {
			s.errorChan <- err
		}
	}()

	s.monitoring = true

	for {
		select {
		case <-s.ctx.Done():
			s.log.Info().Msg("Network polling stopped")
			ticker.Stop()
			return s.ctx.Err()
		case <-s.pauseChan:
			s.pauseChan <- struct{}{}
			return nil
		case r := <-s.resultChan:
			switch r.Type {
			case scanner.ARPResult:
				res := r.Payload.(*scanner.ArpScanResult)
				dr := &DiscoveryResult{
					Type:     DiscoveryArpUpdateEvent,
					ID:       res.MAC.String(),
					IP:       res.IP.String(),
					Hostname: "Unknown",
					OS:       "Unknown",
					Vendor:   res.Vendor,
					Status:   ServerOnline,
					Port: Port{
						ID:     22,
						Status: PortClosed,
					},
				}
				go s.handleArpDiscoveryResult(dr)
			case scanner.SYNResult:
				res := r.Payload.(*scanner.SynScanResult)
				dr := &DiscoveryResult{
					Type:     DiscoverySynUpdateEvent,
					ID:       res.MAC.String(),
					IP:       res.IP.String(),
					Hostname: "",
					OS:       "",
					Vendor:   "",
					Status:   ServerStatus(res.Status),
					Port: Port{
						ID:     res.Port.ID,
						Status: PortStatus(res.Port.Status),
					},
				}
				go s.handleSynDiscoveryResult(dr)
			}
		case err := <-s.errorChan:
			s.log.Error().Err(err).Msg("discovery service encountered an error")
			s.eventManager.SendFatalError(err)
			return err
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

func (s *ScannerService) getConfiguredSSHPort(result *DiscoveryResult) *string {
	resultStrPort := strconv.Itoa(int(result.Port.ID))
	sshPort := s.conf.SSH.Port

	for _, o := range s.conf.SSH.Overrides {
		if result.IP == o.Target {
			if o.Port != "" {
				sshPort = o.Port
			}
			break
		}
	}

	if resultStrPort != sshPort {
		return nil
	}

	return &sshPort
}

func (s *ScannerService) handleArpDiscoveryResult(result *DiscoveryResult) {
	if result.Type != DiscoveryArpUpdateEvent {
		return
	}

	fields := map[string]interface{}{
		"type":     result.Type,
		"id":       result.ID,
		"hostname": result.Hostname,
		"ip":       result.IP,
		"os":       result.OS,
		"status":   result.Status,
	}

	s.log.Info().Fields(fields).Msg("found network device")

	s.eventManager.Send(
		event.Event{
			Type:    event.EventType(result.Type),
			Payload: *result,
		},
	)
}

// handle results found during polling
func (s *ScannerService) handleSynDiscoveryResult(result *DiscoveryResult) {
	if result.Type != DiscoverySynUpdateEvent {
		return
	}

	fields := map[string]interface{}{
		"type":      result.Type,
		"id":        result.ID,
		"hostname":  result.Hostname,
		"ip":        result.IP,
		"os":        result.OS,
		"status":    result.Status,
		"sshPort":   result.Port.ID,
		"sshStatus": result.Port.Status,
	}

	s.log.Info().Fields(fields).Msg("found network device")

	sshPort := s.getConfiguredSSHPort(result)

	if sshPort == nil {
		s.log.Info().Fields(fields).Msg("ignoring non-ssh port result")
		return
	}

	if result.Port.Status == PortOpen {
		details, err := s.detailScanner.GetServerDetails(
			s.ctx,
			result.IP,
			*sshPort,
		)

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
		result.Hostname = "Unknown"
	}

	if result.OS == "" {
		result.OS = "Unknown"
	}

	s.eventManager.Send(
		event.Event{
			Type:    event.EventType(result.Type),
			Payload: *result,
		},
	)
}

func (s *ScannerService) pause() {
	s.pauseChan <- struct{}{}
	<-s.pauseChan
}
