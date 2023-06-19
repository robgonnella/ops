package config_test

import (
	"os"
	"testing"

	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/exception"
	"github.com/robgonnella/opi/internal/test_util"
	"github.com/stretchr/testify/assert"
)

func TestConfigSqliteRepo(t *testing.T) {
	testDBFile := "config.db"

	defer func() {
		os.RemoveAll(testDBFile)
	}()

	db, err := test_util.GetDBConnection(testDBFile)

	if err != nil {
		t.Logf("failed to create test db: %s", err.Error())
		t.FailNow()
	}

	if err := test_util.Migrate(db, config.ConfigModel{}); err != nil {
		t.Logf("failed to migrate test db: %s", err.Error())
		t.FailNow()
	}

	repo := config.NewSqliteRepo(db)

	t.Run("returns record not found error", func(st *testing.T) {
		_, err := repo.Get("noop")

		assert.Error(st, err)
		assert.Equal(st, exception.ErrRecordNotFound, err)
	})

	t.Run("creates, reads, updates, and destroys config", func(st *testing.T) {
		conf := &config.Config{
			Name: "test",
			SSH: config.SSHConfig{
				User:     "test-user",
				Identity: "test-identity",
				Overrides: []config.SSHOverride{
					{
						Target:   "test-target",
						User:     "user",
						Identity: "identity",
					},
				},
			},
			Targets: []string{"target"},
		}

		newConf, err := repo.Create(conf)

		assert.NoError(st, err)
		assert.Equal(st, conf, newConf)

		foundConf, err := repo.Get(newConf.Name)

		assert.NoError(st, err)
		assert.Equal(st, newConf, foundConf)

		conf.SSH.User = "newUser"

		updatedConf, err := repo.Update(conf)

		assert.NoError(st, err)
		assert.Equal(st, "newUser", updatedConf.SSH.User)

		err = repo.Delete(conf.Name)

		assert.NoError(st, err)

		deletedConfig, err := repo.Get(conf.Name)

		assert.Error(st, err)
		assert.Equal(st, exception.ErrRecordNotFound, err)
		assert.Nil(st, deletedConfig)
	})

	t.Run("gets all configs and gets last loaded", func(st *testing.T) {
		conf1 := &config.Config{
			Name: "conf1",
			SSH: config.SSHConfig{
				User:     "test-user1",
				Identity: "test-identity1",
			},
			Targets: []string{"target1"},
		}

		conf2 := &config.Config{
			Name: "conf2",
			SSH: config.SSHConfig{
				User:     "test-user2",
				Identity: "test-identity2",
			},
			Targets: []string{"target2"},
		}

		_, err := repo.Create(conf1)

		assert.NoError(st, err)

		_, err = repo.Create(conf2)

		assert.NoError(st, err)

		confs, err := repo.GetAll()

		assert.NoError(st, err)
		assert.Equal(st, 2, len(confs))

		for _, c := range confs {
			if c.Name == "conf1" {
				assert.Equal(st, conf1, c)
			} else {
				assert.Equal(st, conf2, c)
			}
		}

		lastLoaded, err := repo.LastLoaded()

		assert.NoError(st, err)
		assert.Equal(st, conf2, lastLoaded)
	})
}