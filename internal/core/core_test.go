package core_test

import (
	"context"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/core"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/event"
	mock_config "github.com/robgonnella/ops/internal/mock/config"
	mock_discovery "github.com/robgonnella/ops/internal/mock/discovery"
	mock_server "github.com/robgonnella/ops/internal/mock/server"
	"github.com/robgonnella/ops/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestCore(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockScanner := mock_discovery.NewMockScanner(ctrl)
	mockDetailsScanner := mock_discovery.NewMockDetailScanner(ctrl)
	mockConfig := mock_config.NewMockService(ctrl)
	mockServerService := mock_server.NewMockService(ctrl)

	discoveryService := discovery.NewScannerService(
		mockScanner,
		mockDetailsScanner,
		mockServerService,
	)

	conf := config.Config{
		ID:   1,
		Name: "default",
		SSH: config.SSHConfig{
			User:     "user",
			Identity: "identity",
		},
		Targets: []string{"172.100.1.1/24"},
	}

	coreService := core.New(
		&conf,
		mockConfig,
		mockServerService,
		discoveryService,
	)

	t.Run("returns config", func(st *testing.T) {
		c := coreService.Conf()

		assert.Equal(st, conf, c)
	})

	t.Run("updates config", func(st *testing.T) {
		defer coreService.UpdateConfig(conf)

		newConf := config.Config{
			ID:   1,
			Name: "new",
			SSH: config.SSHConfig{
				User:     "new-user",
				Identity: "new-identity",
			},
			Targets: []string{"new-target"},
		}

		mockConfig.EXPECT().Update(&newConf).Return(&newConf, nil)
		mockConfig.EXPECT().Update(&conf).Return(&conf, nil)
		mockConfig.EXPECT().SetLastLoaded(newConf.ID)
		mockConfig.EXPECT().SetLastLoaded(conf.ID)

		err := coreService.UpdateConfig(newConf)

		assert.NoError(st, err)
		assert.Equal(st, coreService.Conf(), newConf)
	})

	t.Run("sets config", func(st *testing.T) {
		defer coreService.SetConfig(conf.ID)

		anotherConf := config.Config{
			ID:   2,
			Name: "other-conf",
			SSH: config.SSHConfig{
				User:     "other-user",
				Identity: "other-identity",
			},
			Targets: []string{"other target"},
		}

		mockConfig.EXPECT().Get(anotherConf.ID).Return(&anotherConf, nil)
		mockConfig.EXPECT().Get(conf.ID).Return(&conf, nil)
		mockConfig.EXPECT().SetLastLoaded(anotherConf.ID)
		mockConfig.EXPECT().SetLastLoaded(conf.ID)

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
			Targets: []string{"new-target"},
		}

		mockConfig.EXPECT().Create(&newConf).Return(&newConf, nil)

		err := coreService.CreateConfig(newConf)

		assert.NoError(st, err)
		// does not update the "set" config in core
		assert.Equal(st, coreService.Conf(), conf)
	})

	t.Run("deletes config", func(st *testing.T) {
		mockConfig.EXPECT().Delete(10).Return(nil)

		err := coreService.DeleteConfig(10)

		assert.NoError(st, err)
	})

	t.Run("gets all configs", func(st *testing.T) {
		anotherConf := config.Config{
			Name: "other-conf",
			SSH: config.SSHConfig{
				User:     "other-user",
				Identity: "other-identity",
			},
			Targets: []string{"other target"},
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

	t.Run("registers and removes event listener", func(st *testing.T) {
		evtChan := make(chan *event.Event)
		id := coreService.RegisterEventListener(evtChan)

		assert.Equal(st, 1, id)

		coreService.RemoveEventListener(id)
	})

	t.Run("registers and removes server listener", func(st *testing.T) {
		serverChan := make(chan []*server.Server)

		id := coreService.RegisterServerPollListener(serverChan)

		assert.Equal(st, 2, id)

		coreService.RemoveServerPollListener(id)
	})

	t.Run("monitors network", func(st *testing.T) {
		discoveryResults := []*discovery.DiscoveryResult{
			{
				ID:       "id",
				Hostname: "hostname",
				IP:       "ip",
				OS:       "os",
				Status:   server.StatusOnline,
				Ports: []discovery.Port{
					{
						ID:     22,
						Status: discovery.PortOpen,
					},
				},
			},
		}

		details := &discovery.Details{
			Hostname: "hostname",
			OS:       "os",
		}

		serverToUpdate := &server.Server{
			ID:        discoveryResults[0].ID,
			Hostname:  details.Hostname,
			IP:        discoveryResults[0].IP,
			OS:        details.OS,
			Status:    discoveryResults[0].Status,
			SshStatus: server.SSHEnabled,
		}

		wg := sync.WaitGroup{}
		wg.Add(5)

		mockServerService.EXPECT().StreamEvents(gomock.Any()).Return(1)
		mockServerService.EXPECT().
			GetAllServersInNetworkTargets(conf.Targets).
			Do(func([]string) {
				wg.Done()
			})
		mockScanner.EXPECT().Scan().DoAndReturn(func() ([]*discovery.DiscoveryResult, error) {
			defer wg.Done()
			return discoveryResults, nil
		})
		mockDetailsScanner.EXPECT().
			GetServerDetails(gomock.Any(), "ip").
			DoAndReturn(func(
				ctx context.Context,
				ip string,
			) (*discovery.Details, error) {
				defer wg.Done()
				return details, nil
			})
		mockServerService.EXPECT().AddOrUpdateServer(serverToUpdate).Do(
			func(*server.Server) {
				coreService.Stop()
				wg.Done()
			},
		)
		mockScanner.EXPECT().Stop()
		mockServerService.EXPECT().StopStream(1).Do(func(int) {
			wg.Done()
		})

		go coreService.Monitor()

		wg.Wait()
	})
}
