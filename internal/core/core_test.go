package core_test

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/robgonnella/go-lanscan/pkg/scanner"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/core"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	mock_config "github.com/robgonnella/ops/internal/mock/config"
	mock_discovery "github.com/robgonnella/ops/internal/mock/discovery"
	mock_event "github.com/robgonnella/ops/internal/mock/event"
	"github.com/stretchr/testify/assert"
)

type mockNetwork struct {
	iface    *net.Interface
	ipnet    *net.IPNet
	userIP   net.IP
	gateway  net.IP
	cidr     string
	hostname string
}

var mockNet = &mockNetwork{
	iface:    &net.Interface{},
	ipnet:    &net.IPNet{},
	gateway:  net.ParseIP("172.17.0.1"),
	userIP:   net.ParseIP("172.17.0.1"),
	cidr:     "172.17.0.1/32",
	hostname: "mock_hostname",
}

func (n *mockNetwork) Interface() *net.Interface {
	return n.iface
}

func (n *mockNetwork) IPNet() *net.IPNet {
	return n.ipnet
}

func (n *mockNetwork) Gateway() net.IP {
	return n.gateway
}

func (n *mockNetwork) UserIP() net.IP {
	return n.userIP
}

func (n *mockNetwork) Cidr() string {
	return n.cidr
}

func (n *mockNetwork) Hostname() string {
	return n.hostname
}

func TestCore(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockScanner := mock_discovery.NewMockScanner(ctrl)
	mockDetailsScanner := mock_discovery.NewMockDetailScanner(ctrl)
	mockConfig := mock_config.NewMockService(ctrl)
	mockEventManager := mock_event.NewMockManager(ctrl)

	resultChan := make(chan *scanner.ScanResult)

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

	discoveryService := discovery.NewScannerService(
		conf,
		mockScanner,
		mockDetailsScanner,
		resultChan,
		mockEventManager,
	)

	coreService := core.New(
		mockNet,
		&conf,
		mockConfig,
		discoveryService,
		mockEventManager,
		false,
	)

	t.Run("returns config", func(st *testing.T) {
		c := coreService.Conf()

		assert.Equal(st, conf, c)
	})

	t.Run("updates config", func(st *testing.T) {
		defer coreService.UpdateConfig(conf)

		newConf := config.Config{
			ID:   "1",
			Name: "new",
			SSH: config.SSHConfig{
				User:     "new-user",
				Identity: "new-identity",
			},
			CIDR: "192.111.1.1/28",
		}

		mockConfig.EXPECT().Update(&newConf).Return(&newConf, nil)
		mockConfig.EXPECT().Update(&conf).Return(&conf, nil)

		err := coreService.UpdateConfig(newConf)

		assert.NoError(st, err)
		assert.Equal(st, coreService.Conf(), newConf)
	})

	t.Run("sets config", func(st *testing.T) {
		defer coreService.SetConfig(conf.ID)

		anotherConf := config.Config{
			ID:   "2",
			Name: "other-conf",
			SSH: config.SSHConfig{
				User:     "other-user",
				Identity: "other-identity",
			},
			CIDR: "172.22.2.2/32",
		}

		mockConfig.EXPECT().Get(anotherConf.ID).Return(&anotherConf, nil)
		mockConfig.EXPECT().Get(conf.ID).Return(&conf, nil)

		err := coreService.SetConfig(anotherConf.ID)

		assert.NoError(st, err)
		assert.Equal(st, coreService.Conf(), anotherConf)
	})

	t.Run("creates config", func(st *testing.T) {
		newConf := config.Config{
			Name: "new",
			SSH: config.SSHConfig{
				User:     "new-user",
				Identity: "new-identity",
			},
			CIDR: "172.22.2.2/32",
		}

		mockConfig.EXPECT().Create(&newConf).Return(&newConf, nil)

		err := coreService.CreateConfig(newConf)

		assert.NoError(st, err)
		// does not update the "set" config in core
		assert.Equal(st, coreService.Conf(), conf)
	})

	t.Run("deletes config", func(st *testing.T) {
		mockConfig.EXPECT().Delete("10").Return(nil)

		err := coreService.DeleteConfig("10")

		assert.NoError(st, err)
	})

	t.Run("gets all configs", func(st *testing.T) {
		anotherConf := config.Config{
			Name: "other-conf",
			SSH: config.SSHConfig{
				User:     "other-user",
				Identity: "other-identity",
			},
			CIDR: "172.22.2.3/32",
		}

		expectedConfs := []*config.Config{&conf, &anotherConf}

		mockConfig.EXPECT().GetAll().Return(expectedConfs, nil)

		confs, err := coreService.GetConfigs()

		assert.NoError(st, err)
		assert.Equal(st, 2, len(confs))

		for _, c := range confs {
			if c.Name == conf.Name {
				assert.Equal(st, &conf, c)
			} else {
				assert.Equal(st, &anotherConf, c)
			}
		}
	})

	t.Run("monitors network", func(st *testing.T) {
		mac, _ := net.ParseMAC("00:00:00:00:00:00")

		synResults := []*scanner.ScanResult{{
			Type: scanner.SYNResult,
			Payload: &scanner.SynScanResult{
				MAC:    mac,
				IP:     net.ParseIP("127.0.0.1"),
				Status: scanner.StatusOnline,
				Port: scanner.Port{
					ID:      22,
					Service: "",
					Status:  scanner.PortOpen,
				},
			},
		}}

		details := &discovery.Details{
			Hostname: "hostname",
			OS:       "os",
		}

		wg := sync.WaitGroup{}
		wg.Add(3)

		mockScanner.EXPECT().Scan().DoAndReturn(func() error {
			defer wg.Done()
			go func() {
				for _, r := range synResults {
					resultChan <- r
				}
			}()
			return nil
		})

		mockDetailsScanner.EXPECT().
			GetServerDetails(gomock.Any(), "127.0.0.1", conf.SSH.Port).
			DoAndReturn(func(
				ctx context.Context,
				ip string,
				port string,
			) (*discovery.Details, error) {
				defer wg.Done()
				return details, nil
			})

		expectedEvent := event.Event{
			Type: discovery.DiscoverySynUpdateEvent,
			Payload: discovery.DiscoveryResult{
				Type:     discovery.DiscoverySynUpdateEvent,
				ID:       mac.String(),
				Hostname: details.Hostname,
				IP:       "127.0.0.1",
				OS:       details.OS,
				Status:   discovery.ServerOnline,
				Port: discovery.Port{
					ID:     22,
					Status: discovery.PortOpen,
				},
			},
		}

		mockEventManager.EXPECT().Send(expectedEvent).DoAndReturn(func(evt event.Event) {
			wg.Done()
		})

		go coreService.Monitor()

		wg.Wait()
	})
}
