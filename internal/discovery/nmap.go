package discovery

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/Ullaakut/nmap/v3"
	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/logger"
	"github.com/robgonnella/opi/internal/server"
)

// NmapService implements our discovery service using nmap
type NmapService struct {
	ctx            context.Context
	cancel         context.CancelFunc
	conf           config.Config
	serverService  server.Service
	scanner        *nmap.Scanner
	ansibleScanner *AnsibleIpScanner
	logger         logger.Logger
	targets        []string
}

// NewNmapService returns a new intance of nmap network discovery NmapService
func NewNmapService(conf config.Config, serverService server.Service) (*NmapService, error) {
	log := logger.New()

	// Use a cancelable context so we can properly cleanup when needed
	ctxWithCancel, cancel := context.WithCancel(context.Background())

	scanner, err := nmap.NewScanner(
		ctxWithCancel,
		nmap.WithTargets(conf.Discovery.Targets...),
		nmap.WithPorts("22"),
		nmap.WithAggressiveScan(),
		nmap.WithVerbosity(10),
	)

	if err != nil {
		cancel()
		return nil, err
	}

	return &NmapService{
		ctx:            ctxWithCancel,
		cancel:         cancel,
		conf:           conf,
		serverService:  serverService,
		logger:         log,
		scanner:        scanner,
		ansibleScanner: NewAnsibleIpScanner(conf),
		targets:        conf.Discovery.Targets,
	}, nil
}

// SetTargets sets targets for scanner
func (s *NmapService) SetTargets(targets []string) {
	updateScanner := nmap.WithTargets(targets...)
	updateScanner(s.scanner)
}

// MonitorNetwork polls the network and calls out to grpc with the results
func (s *NmapService) MonitorNetwork() {
	s.logger.Info().Msg("Starting network discovery")

	// blocking call that continuously scans the network on an interval
	s.pollNetwork()
}

// Stop stop network discover
func (s *NmapService) Stop() {
	s.cancel()
}

// private
// pollNetwork runs Discover function on an interval to discover devices on the network
func (s *NmapService) pollNetwork() {
	pollTime := time.Second * 30

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info().Msg("Network polling stopped")
			s.cancel()
			return
		default:
			results, err := s.scan()

			if err != nil {
				s.logger.Warn().Err(err).Msg("Error polling network")
			} else {
				s.logger.Info().
					Fields(map[string]interface{}{"count": len(results)}).
					Msg("Discovery results")

				s.handleDiscoveryResults(results)
			}

			time.Sleep(pollTime)
		}
	}
}

// scan targets and ports and return network results
func (s *NmapService) scan() ([]*DiscoveryResult, error) {
	s.logger.Info().Msg("Scanning network...")

	result, warnings, err := s.scanner.Run()

	if len(*warnings) > 0 {
		fields := map[string]interface{}{}

		for i, warning := range *warnings {
			fields[strconv.Itoa(i)] = warning
		}

		s.logger.Warn().
			Fields(fields).
			Msg("encountered network scan warnings")
	}

	if err != nil {
		s.logger.Error().Err(err).Msg("encountered network scan error")
		return nil, err
	}

	discoverResults := []*DiscoveryResult{}

	for _, host := range result.Hosts {
		ports := []Port{}

		for _, port := range host.Ports {
			status := PortClosed

			if port.Status() == nmap.Open {
				status = PortOpen
			}

			ports = append(ports, Port{
				ID:     port.ID,
				Status: status,
			})
		}

		status := server.StatusOffline
		nmapStatus := host.Status

		if nmapStatus.String() == "up" {
			status = server.StatusOnline
		}

		ip := ""

		if len(host.Addresses) > 0 {
			ip = host.Addresses[0].String()
		}

		if ip == "" {
			continue
		}

		hashedIP := sha1.Sum([]byte(ip))
		id := hex.EncodeToString(hashedIP[:])

		res := &DiscoveryResult{
			ID:     id,
			IP:     ip,
			Status: status,
			Ports:  ports,
		}

		discoverResults = append(discoverResults, res)
	}

	return discoverResults, nil
}

func (s *NmapService) handleDiscoveryResults(results []*DiscoveryResult) {
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

		s.logger.Info().Fields(fields).Msg("found network device")

		switch result.Status {
		case server.StatusOnline:
			if deviceType == ServerDevice {
				go s.setServerToOnline(result)
			} else {
				s.logger.Info().
					Str("ip", result.IP).
					Msg("Unknown device detected on network")
			}
		case server.StatusOffline:
			go s.setServerToOffline(result)
		default:
			s.logger.Info().
				Str("status", string(result.Status)).
				Msg("Device detected with unimplemented status action")
		}
	}
}

func (s *NmapService) getDeviceType(result *DiscoveryResult) DeviceType {
	for _, port := range result.Ports {
		if port.ID == 22 {
			return ServerDevice
		}
	}

	return UnknownDevice
}

func (s *NmapService) getSSHStatus(result *DiscoveryResult) server.SSHStatus {
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
func (s *NmapService) setServerToOnline(result *DiscoveryResult) {
	// finally update our server to "online"
	sshStatus := s.getSSHStatus(result)

	if sshStatus == server.SSHEnabled {
		details, err := s.ansibleScanner.GetServerDetails(s.ctx, result.IP)

		if err == nil {
			result.Hostname = details.Hostname
			result.OS = details.OS
		} else {
			s.logger.
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

	s.logger.Info().
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
		s.logger.Error().Err(err).Msg("error marking device online")
	}
}

func (s *NmapService) setServerToOffline(result *DiscoveryResult) {
	s.logger.Info().Str("ip", result.IP).Msg("Marking device offline")
	if err := s.serverService.MarkServerOffline(result.IP); err != nil {
		s.logger.Error().Err(err).Msg("error marking device offline")
	}
}
