package discovery

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/projectdiscovery/mapcidr"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/server"
)

var cidrSuffix = regexp.MustCompile(`\/\d{2}$`)

type NetScanner struct {
	canceled  bool
	targets   []string
	semaphore chan struct{}
	log       logger.Logger
}

func NewNetScanner(targets []string) (*NetScanner, error) {
	ipList := []string{}

	for _, t := range targets {
		if cidrSuffix.MatchString(t) {
			ips, err := mapcidr.IPAddresses(t)

			if err != nil {
				return nil, err
			}

			ipList = append(ipList, ips...)
		} else {
			ipList = append(ipList, t)
		}
	}

	return &NetScanner{
		canceled:  false,
		targets:   ipList,
		semaphore: make(chan struct{}, 1000),
		log:       logger.New(),
	}, nil
}

func (s *NetScanner) Scan(resultChan chan *DiscoveryResult) error {
	if s.canceled {
		return errors.New("network scanner is in a canceled state")
	}

	s.log.Info().Msg("Scanning network...")

	wg := &sync.WaitGroup{}

	for _, ip := range s.targets {
		s.semaphore <- struct{}{} // acquire
		wg.Add(1)
		go func(i string, w *sync.WaitGroup, res chan *DiscoveryResult) {
			r := s.scanIP(i)
			res <- r
			<-s.semaphore // release
			w.Done()
		}(ip, wg, resultChan)
	}

	wg.Wait()

	return nil
}

func (s *NetScanner) Stop() {
	s.canceled = true
}

func (s *NetScanner) scanIP(ip string) *DiscoveryResult {
	hashedIP := sha1.Sum([]byte(ip))
	id := hex.EncodeToString(hashedIP[:])

	r := DiscoveryResult{
		ID:       id,
		Hostname: "",
		IP:       ip,
		OS:       "",
		Status:   server.StatusOffline,
	}

	s.log.Debug().Str("ip", ip).Msg("Scanning target")

	timeOut := time.Millisecond * 200
	conn, err := net.DialTimeout("tcp", ip+":22", timeOut)

	if err != nil {
		r.Ports = []Port{{ID: 22, Status: PortClosed}}

		if _, ok := err.(*net.OpError); ok {
			if strings.HasSuffix(err.Error(), "connect: connection refused") {
				r.Status = server.StatusOnline
			}
		}
	} else {
		defer conn.Close()
		r.Status = server.StatusOnline
		r.Ports = []Port{{ID: 22, Status: PortOpen}}
	}

	return &r
}
