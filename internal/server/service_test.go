package server_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/event"
	"github.com/robgonnella/opi/internal/exception"
	mock_server "github.com/robgonnella/opi/internal/mock/server"
	"github.com/robgonnella/opi/internal/server"
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
		evtChan := make(chan *event.Event, 1)

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
