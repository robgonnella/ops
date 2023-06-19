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
	User      string        `json:"user"`
	Identity  string        `json:"identity"`
	Overrides []SSHOverride `json:"overrides"`
}

// Config represents the data structure of our user provided json configuration
type Config struct {
	Name    string    `json:"name"`
	SSH     SSHConfig `json:"ssh"`
	Targets []string  `json:"targets"`
}

type ConfigModel struct {
	Name   string `gorm:"primaryKey"`
	Data   datatypes.JSON
	Loaded time.Time
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
