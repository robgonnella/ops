package discovery_test

import (
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/robgonnella/go-lanscan/scanner"
	"github.com/robgonnella/ops/internal/discovery"
	mock_discovery "github.com/robgonnella/ops/internal/mock/discovery"
	mock_server "github.com/robgonnella/ops/internal/mock/server"
	"github.com/robgonnella/ops/internal/server"
)

func TestDiscoveryService(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	t.Run("monitors network for offline servers", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		mockServerService := mock_server.NewMockService(ctrl)
		resultChan := make(chan *scanner.SynScanResult)
		doneChan := make(chan bool)

		service := discovery.NewScannerService(
			mockScanner,
			mockDetailScanner,
			mockServerService,
			resultChan,
			doneChan,
		)

		port := scanner.Port{
			ID:      22,
			Service: "ssh",
			Status:  scanner.PortClosed,
		}

		mac, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")

		result := &scanner.SynScanResult{
			MAC:    mac,
			IP:     net.ParseIP("127.0.0.1"),
			Status: scanner.StatusOffline,
			Port:   port,
		}

		mockScanner.EXPECT().Scan().DoAndReturn(func() error {
			go func() {
				resultChan <- result
			}()
			return nil
		})
		mockScanner.EXPECT().Stop()
		mockServerService.EXPECT().MarkServerOffline(result.IP.String())

		go service.MonitorNetwork()

		time.Sleep(time.Millisecond * 10)

		service.Stop()
	})

	t.Run("monitors network for online servers", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		mockServerService := mock_server.NewMockService(ctrl)
		resultChan := make(chan *scanner.SynScanResult)
		doneChan := make(chan bool)

		mac, _ := net.ParseMAC("00:00:00:00:00:00")

		service := discovery.NewScannerService(
			mockScanner,
			mockDetailScanner,
			mockServerService,
			resultChan,
			doneChan,
		)

		port := scanner.Port{
			ID:      22,
			Service: "ssh",
			Status:  scanner.PortClosed,
		}

		result := &scanner.SynScanResult{
			MAC:    mac,
			IP:     net.ParseIP("127.0.0.1"),
			Status: scanner.StatusOnline,
			Port:   port,
		}

		expectedServerCall := &server.Server{
			ID:        result.MAC.String(),
			Hostname:  "unknown",
			IP:        result.IP.String(),
			OS:        "unknown",
			Status:    server.StatusOnline,
			SshStatus: server.SSHDisabled,
		}

		mockScanner.EXPECT().Scan().DoAndReturn(func() error {
			go func() {
				resultChan <- result
			}()
			return nil
		})
		mockScanner.EXPECT().Stop()
		mockServerService.EXPECT().AddOrUpdateServer(expectedServerCall)

		go service.MonitorNetwork()

		time.Sleep(time.Millisecond * 10)

		service.Stop()
	})

	t.Run("requests extra details when ssh is enabled", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		mockServerService := mock_server.NewMockService(ctrl)
		resultChan := make(chan *scanner.SynScanResult)
		doneChan := make(chan bool)

		mac, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")

		service := discovery.NewScannerService(
			mockScanner,
			mockDetailScanner,
			mockServerService,
			resultChan,
			doneChan,
		)

		port := scanner.Port{
			ID:      22,
			Service: "ssh",
			Status:  scanner.PortOpen,
		}

		result := &scanner.SynScanResult{
			MAC:    mac,
			IP:     net.ParseIP("127.0.0.1"),
			Status: scanner.StatusOnline,
			Port:   port,
		}

		expectedDetails := &discovery.Details{
			Hostname: "fancy-hostname",
			OS:       "fancy-os",
		}

		expectedServerCall := &server.Server{
			ID:        result.MAC.String(),
			Hostname:  expectedDetails.Hostname,
			IP:        result.IP.String(),
			OS:        expectedDetails.OS,
			Status:    server.StatusOnline,
			SshStatus: server.SSHEnabled,
		}

		mockScanner.EXPECT().Scan().DoAndReturn(func() error {
			go func() {
				resultChan <- result
			}()
			return nil
		})
		mockScanner.EXPECT().Stop()
		mockDetailScanner.EXPECT().GetServerDetails(gomock.Any(), result.IP.String()).Return(expectedDetails, nil)
		mockServerService.EXPECT().AddOrUpdateServer(expectedServerCall)

		go service.MonitorNetwork()

		time.Sleep(time.Millisecond * 10)

		service.Stop()
	})

}
