package discovery_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/robgonnella/opi/internal/discovery"
	mock_discovery "github.com/robgonnella/opi/internal/mock/discovery"
	mock_server "github.com/robgonnella/opi/internal/mock/server"
	"github.com/robgonnella/opi/internal/server"
)

func TestDiscoveryService(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	t.Run("monitors network for offline servers", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		mockServerService := mock_server.NewMockService(ctrl)

		service := discovery.NewScannerService(
			mockScanner,
			mockDetailScanner,
			mockServerService,
		)

		port := discovery.Port{
			ID:     22,
			Status: discovery.PortClosed,
		}

		result := &discovery.DiscoveryResult{
			ID:       "id",
			Hostname: "hostname",
			IP:       "ip",
			OS:       "os",
			Status:   server.StatusOffline,
			Ports:    []discovery.Port{port},
		}

		results := []*discovery.DiscoveryResult{result}

		mockScanner.EXPECT().Scan().Return(results, nil)
		mockScanner.EXPECT().Stop()
		mockServerService.EXPECT().MarkServerOffline(result.IP)

		go service.MonitorNetwork()

		time.Sleep(time.Millisecond * 10)

		service.Stop()
	})

	t.Run("monitors network for online servers", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		mockServerService := mock_server.NewMockService(ctrl)

		service := discovery.NewScannerService(
			mockScanner,
			mockDetailScanner,
			mockServerService,
		)

		port := discovery.Port{
			ID:     22,
			Status: discovery.PortClosed,
		}

		result := &discovery.DiscoveryResult{
			ID:       "id",
			Hostname: "hostname",
			IP:       "ip",
			OS:       "os",
			Status:   server.StatusOnline,
			Ports:    []discovery.Port{port},
		}

		results := []*discovery.DiscoveryResult{result}

		expectedServerCall := &server.Server{
			ID:        result.ID,
			Hostname:  result.Hostname,
			IP:        result.IP,
			OS:        result.OS,
			Status:    server.StatusOnline,
			SshStatus: server.SSHDisabled,
		}

		mockScanner.EXPECT().Scan().Return(results, nil)
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

		service := discovery.NewScannerService(
			mockScanner,
			mockDetailScanner,
			mockServerService,
		)

		port := discovery.Port{
			ID:     22,
			Status: discovery.PortOpen,
		}

		result := &discovery.DiscoveryResult{
			ID:       "id",
			Hostname: "hostname",
			IP:       "ip",
			OS:       "os",
			Status:   server.StatusOnline,
			Ports:    []discovery.Port{port},
		}

		results := []*discovery.DiscoveryResult{result}

		expectedDetails := &discovery.Details{
			Hostname: "fancy-hostname",
			OS:       "fancy-os",
		}

		expectedServerCall := &server.Server{
			ID:        result.ID,
			Hostname:  expectedDetails.Hostname,
			IP:        result.IP,
			OS:        expectedDetails.OS,
			Status:    server.StatusOnline,
			SshStatus: server.SSHEnabled,
		}

		mockScanner.EXPECT().Scan().Return(results, nil)
		mockScanner.EXPECT().Stop()
		mockDetailScanner.EXPECT().GetServerDetails(gomock.Any(), result.IP).Return(expectedDetails, nil)
		mockServerService.EXPECT().AddOrUpdateServer(expectedServerCall)

		go service.MonitorNetwork()

		time.Sleep(time.Millisecond * 10)

		service.Stop()
	})

}
