package core

import (
	"errors"
	"time"

	"github.com/robgonnella/go-lanscan/network"
	"github.com/robgonnella/go-lanscan/scanner"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/discovery"
	"github.com/robgonnella/ops/internal/exception"
	"github.com/robgonnella/ops/internal/server"

	"github.com/goombaio/namegenerator"
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
func getDefaultConfig(networkInfo *network.NetworkInfo) *config.Config {
	user := viper.Get("user").(string)
	identity := viper.Get("default-ssh-identity").(string)
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	return &config.Config{
		Name: nameGenerator.Generate(),
		SSH: config.SSHConfig{
			User:      user,
			Identity:  identity,
			Overrides: []config.SSHOverride{},
		},
		CIDR: networkInfo.Cidr,
	}
}

// CreateNewAppCore creates and returns a new instance of *core.Core
func CreateNewAppCore(networkInfo *network.NetworkInfo) (*Core, error) {
	dbFile := viper.Get("database-file").(string)

	db, err := getSqliteDbConnection(dbFile)

	if err != nil {
		return nil, err
	}

	configRepo := config.NewSqliteRepo(db)
	configService := config.NewConfigService(configRepo)

	conf, err := configService.GetByCIDR(networkInfo.Cidr)

	if err != nil {
		if errors.Is(err, exception.ErrRecordNotFound) {
			conf = getDefaultConfig(networkInfo)
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

	resultChan := make(chan *scanner.SynScanResult)
	scanDone := make(chan bool)

	netScanner, err := scanner.NewFullScanner(
		networkInfo,
		[]string{},
		[]string{"22"},
		54321,
		resultChan,
		scanDone,
	)

	if err != nil {
		return nil, err
	}

	detailScanner := discovery.NewAnsibleIpScanner(*conf)

	scannerService := discovery.NewScannerService(
		netScanner,
		detailScanner,
		serverService,
		resultChan,
		scanDone,
	)

	if servers, err := serverService.GetAllServersInNetwork(networkInfo.Cidr); err == nil && len(servers) > 0 {
		for _, s := range servers {
			serverService.MarkServerOffline(s.IP)
		}
	}

	return New(
		networkInfo,
		conf,
		configService,
		serverService,
		scannerService,
	), nil
}
