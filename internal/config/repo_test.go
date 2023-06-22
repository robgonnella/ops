package config_test

import (
	"os"
	"testing"

	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/exception"
	"github.com/robgonnella/ops/internal/test_util"
	"github.com/stretchr/testify/assert"
)

func assertEqualConf(t *testing.T, expected, actual *config.Config) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.SSH.User, actual.SSH.User)
	assert.Equal(t, expected.SSH.Identity, actual.SSH.Identity)

	for i, o := range expected.SSH.Overrides {
		assert.Equal(t, o.Target, actual.SSH.Overrides[i].Target)
		assert.Equal(t, o.User, actual.SSH.Overrides[i].User)
		assert.Equal(t, o.Identity, actual.SSH.Overrides[i].Identity)
	}
}

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
		assertEqualConf(st, conf, newConf)

		foundConf, err := repo.Get(newConf.Name)

		assert.NoError(st, err)
		assertEqualConf(st, newConf, foundConf)

		toUpdate := &config.Config{
			ID: newConf.ID,
			SSH: config.SSHConfig{
				User:      "new-ssh-user",
				Identity:  newConf.SSH.Identity,
				Overrides: newConf.SSH.Overrides,
			},
			Targets: newConf.Targets,
		}

		updatedConf, err := repo.Update(toUpdate)

		assert.NoError(st, err)
		assert.Equal(st, "new-ssh-user", updatedConf.SSH.User)

		err = repo.Delete(conf.Name)

		assert.NoError(st, err)

		deletedConfig, err := repo.Get(conf.Name)

		assert.Error(st, err)
		assert.Equal(st, exception.ErrRecordNotFound, err)
		assert.Nil(st, deletedConfig)
	})

	t.Run("gets all configs", func(st *testing.T) {
		conf1 := &config.Config{
			Name: "test2",
			SSH: config.SSHConfig{
				User:     "test-user1",
				Identity: "test-identity1",
			},
			Targets: []string{"target1"},
		}

		conf2 := &config.Config{
			Name: "test3",
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

		allConfigs, err := repo.GetAll()

		assert.NoError(st, err)

		for _, c := range allConfigs {
			if c.Name == conf1.Name {
				assertEqualConf(st, conf1, c)
			} else if c.Name == conf2.Name {
				assertEqualConf(st, conf2, c)
			}
		}

	})

	t.Run("gets last loaded", func(st *testing.T) {
		conf1 := &config.Config{
			Name: "test4",
			SSH: config.SSHConfig{
				User:     "test-user1",
				Identity: "test-identity1",
			},
			Targets: []string{"target1"},
		}

		conf2 := &config.Config{
			Name: "test5",
			SSH: config.SSHConfig{
				User:     "test-user2",
				Identity: "test-identity2",
			},
			Targets: []string{"target2"},
		}

		newConf1, err := repo.Create(conf1)

		assert.NoError(st, err)

		_, err = repo.Create(conf2)

		assert.NoError(st, err)

		err = repo.SetLastLoaded(newConf1.ID)

		assert.NoError(st, err)

		lastLoaded, err := repo.LastLoaded()

		assert.NoError(st, err)
		assertEqualConf(st, newConf1, lastLoaded)
	})
}
