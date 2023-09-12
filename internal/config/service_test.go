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
			ID:   "1",
			Name: "test",
			SSH: config.SSHConfig{
				User:     "user",
				Identity: "identity",
				Port:     "22",
			},
			CIDR: "172.2.2.2/32",
		}

		mockRepo.EXPECT().Get(expectedConfig.ID).Return(expectedConfig, nil)

		foundConf, err := service.Get("1")

		assert.NoError(st, err)
		assert.Equal(st, expectedConfig, foundConf)
	})

	t.Run("gets all configs", func(st *testing.T) {
		conf1 := &config.Config{
			Name: "test",
			SSH: config.SSHConfig{
				User:     "user",
				Identity: "identity",
				Port:     "22",
			},
			CIDR: "172.2.2.2/32",
		}

		conf2 := &config.Config{
			Name: "test",
			SSH: config.SSHConfig{
				User:     "user",
				Identity: "identity",
				Port:     "22",
			},
			CIDR: "172.2.2.3/32",
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
				Port:     "22",
			},
			CIDR: "172.2.2.2/32",
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
				Port:     "22",
			},
			CIDR: "172.2.2.2/32",
		}

		mockRepo.EXPECT().Update(conf).Return(conf, nil)

		updatedConf, err := service.Update(conf)

		assert.NoError(st, err)
		assert.Equal(st, conf, updatedConf)
	})

	t.Run("deletes config", func(st *testing.T) {
		id := "10"

		mockRepo.EXPECT().Delete(id).Return(nil)

		err := service.Delete(id)

		assert.NoError(st, err)
	})

	t.Run("gets last config by cidr", func(st *testing.T) {
		cidr := "172.2.2.2/32"

		expectedConfig := &config.Config{
			Name: "test",
			SSH: config.SSHConfig{
				User:     "user",
				Identity: "identity",
				Port:     "22",
			},
			CIDR: cidr,
		}

		mockRepo.EXPECT().GetByCIDR(cidr).Return(expectedConfig, nil)

		foundConf, err := service.GetByCIDR(cidr)

		assert.NoError(st, err)
		assert.Equal(st, expectedConfig, foundConf)
	})
}
