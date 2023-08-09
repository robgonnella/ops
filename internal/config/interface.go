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
	User      string
	Identity  string
	Overrides []SSHOverride
}

// Config represents the data structure of our user provided json configuration
type Config struct {
	ID   int
	Name string
	SSH  SSHConfig
	CIDR string
}

// SSHConfigModel represents the ssh config stored in the database
type SSHConfigModel struct {
	User      string
	Identity  string
	Overrides datatypes.JSON
}

// ConfigModel represents the config stored in the database
type ConfigModel struct {
	ID   int            `gorm:"primaryKey"`
	Name string         `gorm:"uniqueIndex"`
	SSH  SSHConfigModel `gorm:"embedded"`
	CIDR string         `gorm:"column:cidr"`
}

// Repo interface representing access to stored configs
type Repo interface {
	Get(id int) (*Config, error)
	GetAll() ([]*Config, error)
	GetByCIDR(cidr string) (*Config, error)
	Create(conf *Config) (*Config, error)
	Update(conf *Config) (*Config, error)
	Delete(id int) error
}

// Service interface for manipulating configurations
type Service interface {
	Get(id int) (*Config, error)
	GetAll() ([]*Config, error)
	GetByCIDR(cidr string) (*Config, error)
	Create(conf *Config) (*Config, error)
	Update(conf *Config) (*Config, error)
	Delete(id int) error
}
