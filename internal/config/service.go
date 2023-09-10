package config

// ConfigService is an implementation of the config.Service interface
type ConfigService struct {
	repo Repo
}

// NewConfigService returns a new instance of ConfigService
func NewConfigService(repo Repo) *ConfigService {
	return &ConfigService{repo: repo}
}

// Get returns a config by id
func (s *ConfigService) Get(id string) (*Config, error) {
	return s.repo.Get(id)
}

// GetAll returns all stored configs
func (s *ConfigService) GetAll() ([]*Config, error) {
	return s.repo.GetAll()
}

func (s *ConfigService) GetByCIDR(cidr string) (*Config, error) {
	return s.repo.GetByCIDR(cidr)
}

// Create creates a new config
func (s *ConfigService) Create(conf *Config) (*Config, error) {
	return s.repo.Create(conf)
}

// Update updates an existing config
func (s *ConfigService) Update(conf *Config) (*Config, error) {
	return s.repo.Update(conf)
}

// Delete deletes a config
func (s *ConfigService) Delete(id string) error {
	return s.repo.Delete(id)
}
