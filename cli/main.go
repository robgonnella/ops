package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/robgonnella/opi/cli/commands"
	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/core"
	"github.com/robgonnella/opi/internal/discovery"
	"github.com/robgonnella/opi/internal/exception"
	"github.com/robgonnella/opi/internal/logger"
	"github.com/robgonnella/opi/internal/name"
	"github.com/robgonnella/opi/internal/server"
	"github.com/robgonnella/opi/internal/ui"
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

// get network interface associated with ip
func getIPNetByIP(ip net.IP) (*net.IPNet, error) {
	interfaces, err := net.Interfaces()

	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()

		if err != nil {
			continue
		}

		for _, addr := range addrs {
			_, ipnet, err := net.ParseCIDR(addr.String())

			if err != nil {
				continue
			}

			if ipnet.Contains(ip) {
				return ipnet, nil
			}
		}
	}

	return nil, errors.New("failed to find IPNet")
}

// get cidr for preferred outbound ip of this machine
func getDefaultCidr() (*string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	ipnet, err := getIPNetByIP(localAddr.IP)

	if err != nil {
		return nil, err
	}

	size, _ := ipnet.Mask.Size()

	cidr := fmt.Sprintf("%s/%d", localAddr.IP, size)

	return &cidr, nil
}

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

func setRuntTimeConfig() error {
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	configDir := path.Join(userHomeDir, ".config", name.APP_NAME)

	if err := os.Mkdir(configDir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	logFile := path.Join(configDir, name.APP_NAME+".log")

	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		return err
	}

	cacheDir := path.Join(userCacheDir, name.APP_NAME)

	if err := os.Mkdir(cacheDir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	dbFile := path.Join(cacheDir, name.APP_NAME+".db")

	defaultCidr, err := getDefaultCidr()

	if err != nil {
		return err
	}

	defaultSSHIdentity := path.Join(userHomeDir, ".ssh", "id_rsa")

	user := os.Getenv("USER")

	// share run-time config globally using viper
	viper.Set("log-file", logFile)
	viper.Set("config-dir", configDir)
	viper.Set("cache-dir", cacheDir)
	viper.Set("database-file", dbFile)
	viper.Set("default-cidr", *defaultCidr)
	viper.Set("default-ssh-identity", defaultSSHIdentity)
	viper.Set("user", user)

	return nil
}

// Entry point for the cli.
func main() {
	log := logger.New()

	err := setRuntTimeConfig()

	if err != nil {
		log.Fatal().Err(err)
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
	appUI := ui.New(appCore)

	// Get the "root" cobra cli command
	cmd := commands.Root(&commands.CommandProps{
		Core: appCore,
		UI:   appUI,
	})

	// Allows "grepping" of command output
	cmd.SetOutput(os.Stdout)

	// execute the cobra command and exit with error code if necessary
	err = cmd.ExecuteContext(context.Background())

	if err != nil {
		log.Fatal().Err(err)
	}
}
