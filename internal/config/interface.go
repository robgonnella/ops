package config

//go:generate mockgen -destination=../mock/config/mock_config.go -package=mock_config . Repo,Service

// SSHOverride represents the config needed to
// override ssh config for a single target
type SSHOverride struct {
	Target   string `json:"target"`
	User     string `json:"user"`
	Identity string `json:"identity"`
	Port     string `json:"port"`
}

// SSHConfig represents the config needed to ssh to servers
type SSHConfig struct {
	User      string        `json:"user"`
	Identity  string        `json:"identity"`
	Port      string        `json:"port"`
	Overrides []SSHOverride `json:"overrides"`
}

// Config represents the data structure of our user provided json configuration
type Config struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SSH       SSHConfig `json:"ssh"`
	Interface string    `json:"interface"`
}

// Configs represents our collection of json configs
type Configs struct {
	Configs []*Config `json:"configs"`
}

// Repo interface representing access to stored configs
type Repo interface {
	Get(id string) (*Config, error)
	GetAll() ([]*Config, error)
	GetByInterface(ifaceName string) (*Config, error)
	Create(conf *Config) (*Config, error)
	Update(conf *Config) (*Config, error)
	Delete(id string) error
}

// Service interface for manipulating configurations
type Service interface {
	Get(id string) (*Config, error)
	GetAll() ([]*Config, error)
	GetByInterface(ifaceName string) (*Config, error)
	Create(conf *Config) (*Config, error)
	Update(conf *Config) (*Config, error)
	Delete(id string) error
}
