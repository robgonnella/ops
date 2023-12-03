package discovery_test

import (
	"net"
	"sync"
	"testing"

	"github.com/robgonnella/go-lanscan/pkg/scanner"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	mock_discovery "github.com/robgonnella/ops/internal/mock/discovery"
	mock_event "github.com/robgonnella/ops/internal/mock/event"
	"go.uber.org/mock/gomock"
)

func TestDiscoveryService(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	conf := config.Config{
		ID:   "1",
		Name: "default",
		SSH: config.SSHConfig{
			User:     "user",
			Identity: "identity",
			Port:     "22",
		},
		CIDR: "172.100.1.1/24",
	}

	t.Run("monitors network for offline servers", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		mockEventManager := mock_event.NewMockManager(ctrl)

		resultChan := make(chan *scanner.ScanResult)

		mockScanner.EXPECT().Results().Return(resultChan).AnyTimes()

		service := discovery.NewScannerService(
			conf,
			mockScanner,
			mockDetailScanner,
			mockEventManager,
		)

		port := scanner.Port{
			ID:      22,
			Service: "ssh",
			Status:  scanner.PortClosed,
		}

		mac, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")

		result := &scanner.ScanResult{
			Type: scanner.SYNResult,
			Payload: &scanner.SynScanResult{
				MAC:    mac,
				IP:     net.ParseIP("127.0.0.1"),
				Status: scanner.StatusOffline,
				Port:   port,
			},
		}

		mockScanner.EXPECT().Scan().DoAndReturn(func() error {
			go func() {
				resultChan <- result
			}()
			return nil
		})
		mockScanner.EXPECT().Stop()

		wg := sync.WaitGroup{}
		wg.Add(1)

		expectedEvt := event.Event{
			Type: discovery.DiscoverySynUpdateEvent,
			Payload: discovery.DiscoveryResult{
				Type:     discovery.DiscoverySynUpdateEvent,
				ID:       mac.String(),
				Hostname: "Unknown",
				IP:       "127.0.0.1",
				OS:       "Unknown",
				Status:   discovery.ServerOffline,
				Port: discovery.Port{
					ID:     port.ID,
					Status: discovery.PortClosed,
				},
			},
		}

		mockEventManager.EXPECT().Send(expectedEvt).DoAndReturn(func(evt event.Event) {
			wg.Done()
		})

		go service.MonitorNetwork()

		wg.Wait()

		service.Stop()
	})

	t.Run("monitors network for online servers", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		mockEventManager := mock_event.NewMockManager(ctrl)

		resultChan := make(chan *scanner.ScanResult)

		mockScanner.EXPECT().Results().Return(resultChan).AnyTimes()

		mac, _ := net.ParseMAC("00:00:00:00:00:00")

		service := discovery.NewScannerService(
			conf,
			mockScanner,
			mockDetailScanner,
			mockEventManager,
		)

		port := scanner.Port{
			ID:      22,
			Service: "ssh",
			Status:  scanner.PortClosed,
		}

		result := &scanner.ScanResult{
			Type: scanner.SYNResult,
			Payload: &scanner.SynScanResult{
				MAC:    mac,
				IP:     net.ParseIP("127.0.0.1"),
				Status: scanner.StatusOnline,
				Port:   port,
			},
		}

		mockScanner.EXPECT().Scan().DoAndReturn(func() error {
			go func() {
				resultChan <- result
			}()
			return nil
		})

		mockScanner.EXPECT().Stop()

		wg := sync.WaitGroup{}
		wg.Add(1)

		expectedEvt := event.Event{
			Type: discovery.DiscoverySynUpdateEvent,
			Payload: discovery.DiscoveryResult{
				Type:     discovery.DiscoverySynUpdateEvent,
				ID:       mac.String(),
				Hostname: "Unknown",
				IP:       "127.0.0.1",
				OS:       "Unknown",
				Status:   discovery.ServerOnline,
				Port: discovery.Port{
					ID:     port.ID,
					Status: discovery.PortClosed,
				},
			},
		}

		mockEventManager.EXPECT().Send(expectedEvt).DoAndReturn(func(evt event.Event) {
			wg.Done()
		})

		go service.MonitorNetwork()

		wg.Wait()

		service.Stop()
	})

	t.Run("requests extra details when ssh is enabled", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		mockEventManager := mock_event.NewMockManager(ctrl)

		resultChan := make(chan *scanner.ScanResult)

		mockScanner.EXPECT().Results().Return(resultChan).AnyTimes()

		mac, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")

		service := discovery.NewScannerService(
			conf,
			mockScanner,
			mockDetailScanner,
			mockEventManager,
		)

		port := scanner.Port{
			ID:      22,
			Service: "ssh",
			Status:  scanner.PortOpen,
		}

		resultPayload := &scanner.SynScanResult{
			MAC:    mac,
			IP:     net.ParseIP("127.0.0.1"),
			Status: scanner.StatusOnline,
			Port:   port,
		}

		result := &scanner.ScanResult{
			Type:    scanner.SYNResult,
			Payload: resultPayload,
		}

		expectedDetails := &discovery.Details{
			Hostname: "fancy-hostname",
			OS:       "fancy-os",
		}

		mockScanner.EXPECT().Scan().DoAndReturn(func() error {
			go func() {
				resultChan <- result
			}()
			return nil
		})

		mockDetailScanner.EXPECT().GetServerDetails(gomock.Any(), resultPayload.IP.String(), conf.SSH.Port).Return(expectedDetails, nil)

		mockScanner.EXPECT().Stop()

		expectedEvt := event.Event{
			Type: discovery.DiscoverySynUpdateEvent,
			Payload: discovery.DiscoveryResult{
				Type:     discovery.DiscoverySynUpdateEvent,
				ID:       mac.String(),
				Hostname: "fancy-hostname",
				IP:       "127.0.0.1",
				OS:       "fancy-os",
				Status:   discovery.ServerOnline,
				Port: discovery.Port{
					ID:     port.ID,
					Status: discovery.PortOpen,
				},
			},
		}

		wg := sync.WaitGroup{}
		wg.Add(1)

		mockEventManager.EXPECT().Send(expectedEvt).DoAndReturn(func(evt event.Event) {
			wg.Done()
		})

		go service.MonitorNetwork()

		wg.Wait()

		service.Stop()
	})

}
