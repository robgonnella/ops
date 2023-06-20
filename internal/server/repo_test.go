package server_test

import (
	"os"
	"testing"

	"github.com/robgonnella/ops/internal/exception"
	"github.com/robgonnella/ops/internal/server"
	"github.com/robgonnella/ops/internal/test_util"
	"github.com/stretchr/testify/assert"
)

func TestServerSqliteRepo(t *testing.T) {
	testDBFile := "server.db"

	defer func() {
		os.RemoveAll(testDBFile)
	}()

	db, err := test_util.GetDBConnection(testDBFile)

	if err != nil {
		t.Logf("failed to create test db: %s", err.Error())
		t.FailNow()
	}

	if err := test_util.Migrate(db, server.Server{}); err != nil {
		t.Logf("failed to migrate test db: %s", err.Error())
		t.FailNow()
	}

	repo := server.NewSqliteRepo(db)

	newServer := &server.Server{
		ID:        "id",
		Status:    server.StatusOnline,
		Hostname:  "hostname",
		IP:        "ip",
		OS:        "os",
		SshStatus: server.SSHEnabled,
	}

	t.Run("GetServerByID returns record not found error", func(st *testing.T) {
		_, err := repo.GetServerByID("noop")

		assert.Error(st, err)
		assert.Equal(st, exception.ErrRecordNotFound, err)
	})

	t.Run("GetServerByIP returns record not found error", func(st *testing.T) {
		_, err := repo.GetServerByIP("noop")

		assert.Error(st, err)
		assert.Equal(st, exception.ErrRecordNotFound, err)
	})

	t.Run("adds server", func(st *testing.T) {
		createdServer, err := repo.AddServer(newServer)

		assert.NoError(st, err)
		assert.Equal(st, newServer, createdServer)
	})

	t.Run("gets server by id", func(st *testing.T) {
		id := "id"

		foundServer, err := repo.GetServerByID(id)

		assert.NoError(st, err)
		assert.Equal(st, newServer, foundServer)
	})

	t.Run("gets server by ip", func(st *testing.T) {
		ip := "ip"

		foundServer, err := repo.GetServerByIP(ip)

		assert.NoError(st, err)
		assert.Equal(st, newServer, foundServer)
	})

	t.Run("gets all servers", func(st *testing.T) {
		foundServers, err := repo.GetAllServers()

		assert.NoError(st, err)
		assert.Equal(st, 1, len(foundServers))
		assert.Equal(st, newServer, foundServers[0])
	})

	t.Run("updates server", func(st *testing.T) {
		toUpdate := &server.Server{
			ID:        newServer.ID,
			Hostname:  "new hostname",
			Status:    newServer.Status,
			IP:        newServer.IP,
			OS:        newServer.OS,
			SshStatus: newServer.SshStatus,
		}

		updatedServer, err := repo.UpdateServer(toUpdate)

		assert.NoError(st, err)
		assert.Equal(st, "new hostname", updatedServer.Hostname)
	})

	t.Run("removes server", func(st *testing.T) {
		err := repo.RemoveServer(newServer.ID)

		assert.NoError(st, err)

		foundServer, err := repo.GetServerByID(newServer.ID)

		assert.Nil(st, foundServer)
		assert.Error(st, err)
		assert.Equal(st, exception.ErrRecordNotFound, err)
	})
}
