package config

type ConfigService struct {
	repo Repo
}

func NewConfigService(repo Repo) *ConfigService {
	return &ConfigService{repo: repo}
}

func (s *ConfigService) Get(name string) (*Config, error) {
	return s.repo.Get(name)
}

func (s *ConfigService) Create(conf *Config) (*Config, error) {
	return s.repo.Create(conf)
}

func (s *ConfigService) Update(conf *Config) (*Config, error) {
	return s.repo.Update(conf)
}

func (s *ConfigService) LastLoaded() (*Config, error) {
	return s.repo.LastLoaded()
}
