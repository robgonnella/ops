package config

import (
	"os"
	"path"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// SSHOverride represents the config needed to
// override ssh config for a single target
type SSHOverride struct {
	Target   string `yaml:"target"`
	User     string `yaml:"user"`
	Identity string `yaml:"identity"`
}

// SSHConfig represents the config needed to ssh to servers
type SSHConfig struct {
	User      string        `yaml:"user"`
	Identity  string        `yaml:"identity"`
	Overrides []SSHOverride `yaml:"overrides"`
}

// Discovery represents our network discovery service configuration
type Discovery struct {
	SSH     SSHConfig `yaml:"ssh"`
	Targets []string  `yaml:"targets"`
}

// Config represents the data structure of our user provided yaml configuration
type Config struct {
	Discovery Discovery `yaml:"discovery"`
}

// New returns umarshaled data structure of user provided config
func New(confPath string) (*Config, error) {
	var config Config

	raw, err := os.ReadFile(confPath)

	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(raw, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func Default() (*Config, error) {
	user := os.Getenv("USER")
	home, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	identity := path.Join(home, ".ssh/id_rsa")

	return &Config{
		Discovery: Discovery{
			SSH: SSHConfig{
				User:      user,
				Identity:  identity,
				Overrides: []SSHOverride{},
			},
			Targets: []string{},
		},
	}, err
}

func Write(conf Config) error {
	configFile := viper.Get("config-file").(string)

	file, err := os.Create(configFile)

	if err != nil {
		return err
	}

	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)

	return encoder.Encode(conf)
}
