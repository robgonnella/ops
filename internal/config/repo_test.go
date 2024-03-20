package config_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/exception"
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

func TestConfigJsonRepo(t *testing.T) {
	testConfigFile := "config.json"
	file, err := os.Create(testConfigFile)

	assert.NoError(t, err)

	defer func() {
		os.RemoveAll(testConfigFile)
	}()

	defaultConf := config.Config{
		ID:   "0",
		Name: "default",
		SSH: config.SSHConfig{
			User:      "user",
			Identity:  "./id_rsa",
			Port:      "22",
			Overrides: []config.SSHOverride{},
		},
	}

	conf := config.Config{
		ID:   "1",
		Name: "myConfig",
		SSH: config.SSHConfig{
			User:      "user",
			Identity:  "./id_rsa",
			Port:      "22",
			Overrides: []config.SSHOverride{},
		},
	}

	data, err := json.Marshal(conf)

	assert.NoError(t, err)

	_, err = file.WriteString(string(data))
	file.Close()

	assert.NoError(t, err)

	repo, err := config.NewJSONRepo(testConfigFile, defaultConf)

	assert.NoError(t, err)

	t.Run("returns record not found error", func(st *testing.T) {
		_, err := repo.Get("10")

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
			Interface: "test",
		}

		newConf, err := repo.Create(conf)

		assert.NoError(st, err)
		assertEqualConf(st, conf, newConf)

		foundConf, err := repo.Get(newConf.ID)

		assert.NoError(st, err)
		assertEqualConf(st, newConf, foundConf)

		toUpdate := &config.Config{
			ID: newConf.ID,
			SSH: config.SSHConfig{
				User:      "new-ssh-user",
				Identity:  newConf.SSH.Identity,
				Overrides: newConf.SSH.Overrides,
			},
			Interface: newConf.Interface,
		}

		updatedConf, err := repo.Update(toUpdate)

		assert.NoError(st, err)
		assert.Equal(st, "new-ssh-user", updatedConf.SSH.User)

		err = repo.Delete(updatedConf.ID)

		assert.NoError(st, err)

		deletedConfig, err := repo.Get(updatedConf.ID)

		assert.Error(st, err)
		assert.Equal(st, exception.ErrRecordNotFound, err)
		assert.Nil(st, deletedConfig)
	})

	t.Run("gets all configs", func(st *testing.T) {
		conf1 := &config.Config{
			Name: "test2",
			SSH: config.SSHConfig{
				User:     "test-user2",
				Identity: "test-identity2",
			},
			Interface: "test",
		}

		conf2 := &config.Config{
			Name: "test3",
			SSH: config.SSHConfig{
				User:     "test-user3",
				Identity: "test-identity3",
			},
			Interface: "en1",
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

	t.Run("gets by cidr", func(st *testing.T) {
		conf1 := &config.Config{
			Name: "test4",
			SSH: config.SSHConfig{
				User:     "test-user4",
				Identity: "test-identity4",
			},
			Interface: "test",
		}

		conf2 := &config.Config{
			Name: "test5",
			SSH: config.SSHConfig{
				User:     "test-user5",
				Identity: "test-identity5",
			},
			Interface: "en1",
		}

		_, err := repo.Create(conf1)

		assert.NoError(st, err)

		newConf2, err := repo.Create(conf2)

		assert.NoError(st, err)

		foundConf, err := repo.GetByInterface("en1")

		assert.NoError(st, err)
		assertEqualConf(st, newConf2, foundConf)
	})
}
