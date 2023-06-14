package ui

import (
	"errors"
	"fmt"
	"os"

	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/core"
	"github.com/robgonnella/opi/internal/discovery"
	"github.com/robgonnella/opi/internal/exception"
	"github.com/robgonnella/opi/internal/logger"
	"github.com/robgonnella/opi/internal/server"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// creates default config
func getDefaultConfig() *config.Config {
	user := viper.Get("user").(string)
	identity := viper.Get("default-ssh-identity").(string)
	cidr := viper.Get("default-cidr").(string)

	return &config.Config{
		Name: "default",
		SSH: config.SSHConfig{
			User:      user,
			Identity:  identity,
			Overrides: []config.SSHOverride{},
		},
		Targets: []string{cidr},
	}
}

// get sqlite database connection
func getDbConnection(dbFile string) (*gorm.DB, error) {
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

type UI struct {
	view *view
}

var originalStdout = os.Stdout
var originalStderr = os.Stderr

func New() *UI {
	return &UI{}
}

func (u *UI) Launch() error {
	log := logger.New()

	level := zerolog.GlobalLevel()

	if level != zerolog.Disabled {
		logFile, ok := viper.Get("log-file").(string)

		if !ok || logFile == "" {
			log.Error().Err(
				fmt.Errorf("invalid log file path: %s", logFile),
			)
			log.Info().Msg("disabling logs")
			zerolog.SetGlobalLevel(zerolog.Disabled)
		} else {
			if err := logger.GlobalSetLogFile(logFile); err != nil {
				log.Error().Err(err)
				log.Info().Msg("disabling logs")
				zerolog.SetGlobalLevel(zerolog.Disabled)
			}
		}

	}

	dbFile := viper.Get("database-file").(string)

	db, err := getDbConnection(dbFile)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to open db connection")
	}

	configRepo := config.NewSqliteDatabase(db)
	configService := config.NewConfigService(configRepo)

	conf, err := configService.LastLoaded()

	if err != nil {
		if errors.Is(err, exception.ErrRecordNotFound) {
			conf = getDefaultConfig()
			conf, err = configService.Create(conf)

			if err != nil {
				log.Fatal().Err(err)
			}
		} else {
			log.Fatal().Err(err).Msg("error loading config")
		}
	}

	serverRepo := server.NewSqliteDatabase(db)
	serverService := server.NewService(*conf, serverRepo)

	discoveryService, err := discovery.NewNmapService(
		*conf,
		serverService,
	)

	if err != nil {
		log.Fatal().Err(err)
	}

	appCore := core.New(*conf, configService, serverService, discoveryService)

	u.view = newView(appCore)

	os.Stdout, _ = os.Open(os.DevNull)
	os.Stderr, _ = os.Open(os.DevNull)

	return u.view.run()
}
