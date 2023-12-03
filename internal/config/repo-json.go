package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/robgonnella/ops/internal/exception"
)

// JSONRepo is our repo implementation for json
type JSONRepo struct {
	configPath string
	configs    []*Config
	mux        sync.Mutex
}

// NewJSONRepo returns a new ops repo for flat yaml file
func NewJSONRepo(configPath string) *JSONRepo {
	repo := &JSONRepo{
		configPath: configPath,
		configs:    []*Config{},
		mux:        sync.Mutex{},
	}

	repo.load()

	return repo
}

// Get returns a config from the db
func (r *JSONRepo) Get(id string) (*Config, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if id == "" {
		return nil, errors.New("config id cannot be empty")
	}

	var conf *Config

	for _, c := range r.configs {
		if c.ID == id {
			conf = copyConfig(c)
			break
		}
	}

	if conf == nil {
		return nil, exception.ErrRecordNotFound
	}

	return conf, nil
}

// GetAll returns all configs in db
func (r *JSONRepo) GetAll() ([]*Config, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	return r.configs, nil
}

// Create creates a new config in db
func (r *JSONRepo) Create(conf *Config) (*Config, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	idx := slices.IndexFunc(r.configs, func(c *Config) bool {
		return c.ID == conf.ID
	})

	if idx != -1 {
		return nil, fmt.Errorf("config already exists: ID: %s", conf.ID)
	}

	copy := copyConfig(conf)
	copy.ID = uuid.New().String()

	r.configs = append(r.configs, copy)

	if err := r.write(); err != nil {
		return nil, err
	}

	return copy, nil
}

// Update updates a config in db
func (r *JSONRepo) Update(conf *Config) (*Config, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if conf.ID == "" {
		return nil, errors.New("config ID cannot be empty")
	}

	idx := slices.IndexFunc(r.configs, func(c *Config) bool {
		return c.ID == conf.ID
	})

	if idx == -1 {
		return nil, exception.ErrRecordNotFound
	}

	copy := copyConfig(conf)

	r.configs[idx] = copy

	if err := r.write(); err != nil {
		return nil, err
	}

	return copy, nil
}

// Delete deletes a config from db
func (r *JSONRepo) Delete(id string) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	if id == "" {
		return errors.New("config id cannot be empty")
	}

	configs := []*Config{}

	for _, c := range r.configs {
		if c.ID != id {
			configs = append(configs, c)
		}
	}

	r.configs = configs

	return r.write()
}

// LastLoaded returns the most recently loaded config
func (r *JSONRepo) GetByCIDR(cidr string) (*Config, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	var conf *Config

	for _, c := range r.configs {
		if c.CIDR == cidr {
			conf = copyConfig(c)
		}
	}

	if conf == nil {
		return nil, exception.ErrRecordNotFound
	}

	return conf, nil
}

func (r *JSONRepo) write() error {
	file, err := os.OpenFile(r.configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	configs := Configs{
		Configs: r.configs,
	}

	data, err := json.MarshalIndent(&configs, "", "\t")

	if err != nil {
		return err
	}

	_, err = file.Write(data)

	if err != nil {
		return err
	}

	return r.load()
}

func (r *JSONRepo) load() error {
	file, err := os.Open(r.configPath)

	if err != nil {
		return err
	}

	defer file.Close()

	data, err := io.ReadAll(file)

	if err != nil {
		return err
	}

	configs := Configs{}

	if err := json.Unmarshal(data, &configs); err != nil {
		return err
	}

	r.configs = configs.Configs

	return nil
}

// helpers
func copyConfig(c *Config) *Config {
	return &Config{
		ID:   c.ID,
		Name: c.Name,
		SSH: SSHConfig{
			User:      c.SSH.User,
			Identity:  c.SSH.Identity,
			Port:      c.SSH.Port,
			Overrides: c.SSH.Overrides,
		},
		CIDR: c.CIDR,
	}
}
