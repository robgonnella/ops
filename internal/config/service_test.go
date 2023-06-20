package config_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/robgonnella/ops/internal/config"
	mock_config "github.com/robgonnella/ops/internal/mock/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigService(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockRepo := mock_config.NewMockRepo(ctrl)

	service := config.NewConfigService(mockRepo)

	t.Run("gets config", func(st *testing.T) {
		expectedConfig := &config.Config{
			Name: "test",
			SSH: config.SSHConfig{
				User:     "user",
				Identity: "identity",
			},
			Targets: []string{"target"},
		}

		mockRepo.EXPECT().Get("test").Return(expectedConfig, nil)

		foundConf, err := service.Get("test")

		assert.NoError(st, err)
		assert.Equal(st, expectedConfig, foundConf)
	})

	t.Run("gets all configs", func(st *testing.T) {
		conf1 := &config.Config{
			Name: "test",
			SSH: config.SSHConfig{
				User:     "user",
				Identity: "identity",
			},
			Targets: []string{"target"},
		}

		conf2 := &config.Config{
			Name: "test",
			SSH: config.SSHConfig{
				User:     "user",
				Identity: "identity",
			},
			Targets: []string{"target"},
		}

		expectedConfs := []*config.Config{conf1, conf2}

		mockRepo.EXPECT().GetAll().Return(expectedConfs, nil)

		foundConfs, err := service.GetAll()

		assert.NoError(st, err)
		assert.Equal(st, expectedConfs, foundConfs)
	})

	t.Run("creates config", func(st *testing.T) {
		conf := &config.Config{
			Name: "test",
			SSH: config.SSHConfig{
				User:     "user",
				Identity: "identity",
			},
			Targets: []string{"target"},
		}

		mockRepo.EXPECT().Create(conf).Return(conf, nil)

		createdConf, err := service.Create(conf)

		assert.NoError(st, err)
		assert.Equal(st, conf, createdConf)
	})

	t.Run("creates config", func(st *testing.T) {
		conf := &config.Config{
			Name: "test",
			SSH: config.SSHConfig{
				User:     "user",
				Identity: "identity",
			},
			Targets: []string{"target"},
		}

		mockRepo.EXPECT().Update(conf).Return(conf, nil)

		updatedConf, err := service.Update(conf)

		assert.NoError(st, err)
		assert.Equal(st, conf, updatedConf)
	})

	t.Run("deletes config", func(st *testing.T) {
		name := "test"

		mockRepo.EXPECT().Delete(name).Return(nil)

		err := service.Delete(name)

		assert.NoError(st, err)
	})

	t.Run("gets last loaded config", func(st *testing.T) {
		expectedConfig := &config.Config{
			Name: "test",
			SSH: config.SSHConfig{
				User:     "user",
				Identity: "identity",
			},
			Targets: []string{"target"},
		}

		mockRepo.EXPECT().LastLoaded().Return(expectedConfig, nil)

		foundConf, err := service.LastLoaded()

		assert.NoError(st, err)
		assert.Equal(st, expectedConfig, foundConf)
	})
}
