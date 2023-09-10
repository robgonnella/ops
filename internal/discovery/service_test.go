package discovery_test

import (
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/robgonnella/go-lanscan/scanner"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	mock_discovery "github.com/robgonnella/ops/internal/mock/discovery"
	"github.com/stretchr/testify/assert"
)

func TestDiscoveryService(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	t.Run("monitors network for offline servers", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		resultChan := make(chan *scanner.ScanResult)
		eventChan := make(chan *event.Event)

		service := discovery.NewScannerService(
			mockScanner,
			mockDetailScanner,
			resultChan,
			eventChan,
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

		go service.MonitorNetwork()

		time.Sleep(time.Millisecond * 10)

		service.Stop()

		evt := <-eventChan

		assert.Equal(st, discovery.DiscoverySynUpdateEvent, evt.Type)

		payload, ok := evt.Payload.(*discovery.DiscoveryResult)

		assert.True(st, ok)

		assert.Equal(st, mac.String(), payload.ID)
		assert.Equal(st, discovery.ServerOffline, payload.Status)
	})

	t.Run("monitors network for online servers", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		resultChan := make(chan *scanner.ScanResult)
		eventChan := make(chan *event.Event)

		mac, _ := net.ParseMAC("00:00:00:00:00:00")

		service := discovery.NewScannerService(
			mockScanner,
			mockDetailScanner,
			resultChan,
			eventChan,
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

		go service.MonitorNetwork()

		time.Sleep(time.Millisecond * 10)

		service.Stop()

		evt := <-eventChan

		assert.Equal(st, discovery.DiscoverySynUpdateEvent, evt.Type)

		payload, ok := evt.Payload.(*discovery.DiscoveryResult)

		assert.True(st, ok)

		assert.Equal(st, mac.String(), payload.ID)
		assert.Equal(st, discovery.ServerOnline, payload.Status)
	})

	t.Run("requests extra details when ssh is enabled", func(st *testing.T) {
		mockScanner := mock_discovery.NewMockScanner(ctrl)
		mockDetailScanner := mock_discovery.NewMockDetailScanner(ctrl)
		resultChan := make(chan *scanner.ScanResult)
		eventChan := make(chan *event.Event)

		mac, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")

		service := discovery.NewScannerService(
			mockScanner,
			mockDetailScanner,
			resultChan,
			eventChan,
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
		mockScanner.EXPECT().Stop()
		mockDetailScanner.EXPECT().GetServerDetails(gomock.Any(), resultPayload.IP.String()).Return(expectedDetails, nil)

		go service.MonitorNetwork()

		time.Sleep(time.Millisecond * 10)

		service.Stop()

		evt := <-eventChan

		assert.Equal(st, discovery.DiscoverySynUpdateEvent, evt.Type)

		payload, ok := evt.Payload.(*discovery.DiscoveryResult)

		assert.True(st, ok)

		assert.Equal(st, mac.String(), payload.ID)
		assert.Equal(st, discovery.ServerOnline, payload.Status)
		assert.Equal(st, expectedDetails.Hostname, payload.Hostname)
		assert.Equal(st, expectedDetails.OS, payload.OS)
	})

}
