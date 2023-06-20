package config

import (
	"time"

	"gorm.io/datatypes"
)

//go:generate mockgen -destination=../mock/config/mock_config.go -package=mock_config . Repo,Service

// SSHOverride represents the config needed to
// override ssh config for a single target
type SSHOverride struct {
	Target   string `json:"target"`
	User     string `json:"user"`
	Identity string `json:"identity"`
}

// SSHConfig represents the config needed to ssh to servers
type SSHConfig struct {
	User      string
	Identity  string
	Overrides []SSHOverride
}

// Config represents the data structure of our user provided json configuration
type Config struct {
	ID      int
	Name    string
	SSH     SSHConfig
	Targets []string `json:"targets"`
	Loaded  time.Time
}

// SSHConfigModel represents the ssh config stored in the database
type SSHConfigModel struct {
	User      string
	Identity  string
	Overrides datatypes.JSON
}

// ConfigModel represents the config stored in the database
type ConfigModel struct {
	ID      int            `gorm:"primaryKey"`
	Name    string         `gorm:"uniqueIndex"`
	SSH     SSHConfigModel `gorm:"embedded"`
	Targets datatypes.JSON
	Loaded  time.Time
}

type Repo interface {
	Get(name string) (*Config, error)
	GetAll() ([]*Config, error)
	Create(conf *Config) (*Config, error)
	Update(conf *Config) (*Config, error)
	Delete(name string) error
	LastLoaded() (*Config, error)
}

type Service interface {
	Get(name string) (*Config, error)
	GetAll() ([]*Config, error)
	Create(conf *Config) (*Config, error)
	Update(conf *Config) (*Config, error)
	Delete(name string) error
	LastLoaded() (*Config, error)
}
