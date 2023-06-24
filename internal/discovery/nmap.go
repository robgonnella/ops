package discovery

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strconv"

	"github.com/Ullaakut/nmap/v3"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/server"
)

// NmapScanner is an implementation of the Scanner interface
type NmapScanner struct {
	ctx     context.Context
	cancel  context.CancelFunc
	scanner *nmap.Scanner
	log     logger.Logger
}

// NewNmapScanner returns a new instance of NmapScanner
func NewNmapScanner(targets []string) (*NmapScanner, error) {
	log := logger.New()

	// Use a cancelable context so we can properly cleanup when needed
	ctxWithCancel, cancel := context.WithCancel(context.Background())

	scanner, err := nmap.NewScanner(
		ctxWithCancel,
		nmap.WithTargets(targets...),
		nmap.WithPorts("22"),
		nmap.WithTimingTemplate(nmap.TimingFastest),
		nmap.WithACKDiscovery(),
		nmap.WithVerbosity(10),
	)

	if err != nil {
		cancel()
		return nil, err
	}

	return &NmapScanner{
		ctx:     ctxWithCancel,
		cancel:  cancel,
		log:     log,
		scanner: scanner,
	}, nil
}

// Stop stops network scanning. Once called this scanner will be useless,
// a new one will need to be instantiated to continue scanning.
func (s *NmapScanner) Stop() {
	s.cancel()
}

// scan targets and ports and return network results
func (s *NmapScanner) Scan() ([]*DiscoveryResult, error) {
	s.log.Info().Msg("Scanning network...")

	result, warnings, err := s.scanner.Run()

	if len(*warnings) > 0 {
		fields := map[string]interface{}{}

		for i, warning := range *warnings {
			fields[strconv.Itoa(i)] = warning
		}

		s.log.Warn().
			Fields(fields).
			Msg("encountered network scan warnings")
	}

	if err != nil {
		s.log.Error().Err(err).Msg("encountered network scan error")
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
