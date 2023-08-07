package util

import (
	"errors"

	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/core"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/exception"
	"github.com/robgonnella/ops/internal/server"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// getSqliteDbConnection creates and returns a sqlite database connection
func getSqliteDbConnection(dbFile string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&config.ConfigModel{},
		&server.Server{},
	)

	if err != nil {
		return nil, err
	}

	return db, nil
}

// getDefaultConfig creates and returns a default configuration
func getDefaultConfig(defaultCIDR string) *config.Config {
	user := viper.Get("user").(string)
	identity := viper.Get("default-ssh-identity").(string)

	return &config.Config{
		Name: "default",
		SSH: config.SSHConfig{
			User:      user,
			Identity:  identity,
			Overrides: []config.SSHOverride{},
		},
		Targets: []string{defaultCIDR},
	}
}

// CreateNewAppCore creates and returns a new instance of *core.Core
func CreateNewAppCore(defaultCIDR string) (*core.Core, error) {
	dbFile := viper.Get("database-file").(string)

	db, err := getSqliteDbConnection(dbFile)

	if err != nil {
		return nil, err
	}

	configRepo := config.NewSqliteRepo(db)
	configService := config.NewConfigService(configRepo)

	conf, err := configService.LastLoaded()

	if err != nil {
		if errors.Is(err, exception.ErrRecordNotFound) {
			conf = getDefaultConfig(defaultCIDR)
			conf, err = configService.Create(conf)

			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	serverRepo := server.NewSqliteRepo(db)
	serverService := server.NewService(*conf, serverRepo)

	netScanner, err := discovery.NewNetScanner(conf.Targets)

	if err != nil {
		return nil, err
	}

	detailScanner := discovery.NewAnsibleIpScanner(*conf)

	scannerService := discovery.NewScannerService(
		netScanner,
		detailScanner,
		serverService,
	)

	return core.New(conf, configService, serverService, scannerService), nil
}
