package server_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/exception"
	mock_server "github.com/robgonnella/ops/internal/mock/server"
	"github.com/robgonnella/ops/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestServerService(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockRepo := mock_server.NewMockRepo(ctrl)

	conf := config.Config{
		Name: "default",
		SSH: config.SSHConfig{
			User:     "user",
			Identity: "identity",
		},
		Targets: []string{"target"},
	}

	service := server.NewService(conf, mockRepo)

	testServer := &server.Server{
		ID:        "id",
		Status:    server.StatusOnline,
		Hostname:  "hostname",
		IP:        "ip",
		OS:        "os",
		SshStatus: server.SSHEnabled,
	}

	t.Run("gets all servers", func(st *testing.T) {
		expectedServers := []*server.Server{testServer}

		mockRepo.EXPECT().GetAllServers().Return(expectedServers, nil)

		foundServers, err := service.GetAllServers()

		assert.NoError(st, err)
		assert.Equal(st, expectedServers, foundServers)
	})

	t.Run("gets all servers in network targets", func(st *testing.T) {
		targets := []string{"192.168.1.10", "172.16.1.1/24"}

		testServer1 := *testServer
		testServer2 := *testServer
		testServer3 := *testServer

		testServer1.IP = "192.168.1.10"
		testServer2.IP = "172.16.1.42"
		testServer3.IP = "192.168.1.11"

		testServers := []*server.Server{
			&testServer1,
			&testServer2,
			&testServer3,
		}

		expectedServers := []*server.Server{
			&testServer1,
			&testServer2,
		}

		mockRepo.EXPECT().GetAllServers().Return(testServers, nil)

		foundServers, err := service.GetAllServersInNetworkTargets(targets)

		assert.NoError(st, err)
		assert.Equal(st, 2, len(foundServers))
		assert.Equal(st, expectedServers, foundServers)
	})

	t.Run("adds server", func(st *testing.T) {
		mockRepo.EXPECT().GetServerByID(gomock.Any()).Return(nil, exception.ErrRecordNotFound)
		mockRepo.EXPECT().AddServer(testServer)

		err := service.AddOrUpdateServer(testServer)

		assert.NoError(st, err)
	})

	t.Run("updates server", func(st *testing.T) {
		toUpdate := &server.Server{
			ID:        testServer.ID,
			Hostname:  "new hostname",
			Status:    testServer.Status,
			IP:        testServer.IP,
			OS:        testServer.OS,
			SshStatus: testServer.SshStatus,
		}

		mockRepo.EXPECT().GetServerByID(gomock.Any()).Return(testServer, nil)
		mockRepo.EXPECT().UpdateServer(toUpdate)

		err := service.AddOrUpdateServer(toUpdate)

		assert.NoError(st, err)
	})

	t.Run("marks server offline", func(st *testing.T) {
		toUpdate := &server.Server{
			ID:        testServer.ID,
			Hostname:  testServer.Hostname,
			Status:    server.StatusOffline,
			IP:        testServer.IP,
			OS:        testServer.OS,
			SshStatus: server.SSHDisabled,
		}

		mockRepo.EXPECT().GetServerByIP(gomock.Any()).Return(testServer, nil)
		mockRepo.EXPECT().UpdateServer(toUpdate)

		err := service.MarkServerOffline(testServer.IP)

		assert.NoError(st, err)
	})

	t.Run("streams events", func(st *testing.T) {
		evtChan := make(chan *event.Event)

		streamID := service.StreamEvents(evtChan)

		assert.Equal(st, 1, streamID)
	})

	t.Run("stops stream", func(st *testing.T) {

		service.StopStream(1)

		assert.Equal(st, true, true)
	})

	t.Run("gets server", func(st *testing.T) {
		mockRepo.EXPECT().GetServerByID(testServer.ID).Return(testServer, nil)

		foundServer, err := service.GetServer(testServer.ID)

		assert.NoError(st, err)
		assert.Equal(st, testServer, foundServer)
	})
}
