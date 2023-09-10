package config

import (
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
	ID   string    `json:"id"`
	Name string    `json:"name"`
	SSH  SSHConfig `json:"ssh"`
	CIDR string    `json:"cidr"`
}

type Configs struct {
	Configs []*Config `json:"configs"`
}

// SSHConfigModel represents the ssh config stored in the database
type SSHConfigModel struct {
	User      string
	Identity  string
	Overrides datatypes.JSON
}

// ConfigModel represents the config stored in the database
type ConfigModel struct {
	ID   string         `gorm:"primaryKey"`
	Name string         `gorm:"uniqueIndex"`
	SSH  SSHConfigModel `gorm:"embedded"`
	CIDR string         `gorm:"column:cidr"`
}

// Repo interface representing access to stored configs
type Repo interface {
	Get(id string) (*Config, error)
	GetAll() ([]*Config, error)
	GetByCIDR(cidr string) (*Config, error)
	Create(conf *Config) (*Config, error)
	Update(conf *Config) (*Config, error)
	Delete(id string) error
}

// Service interface for manipulating configurations
type Service interface {
	Get(id string) (*Config, error)
	GetAll() ([]*Config, error)
	GetByCIDR(cidr string) (*Config, error)
	Create(conf *Config) (*Config, error)
	Update(conf *Config) (*Config, error)
	Delete(id string) error
}
