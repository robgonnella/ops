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

func (s *ConfigService) GetAll() ([]*Config, error) {
	return s.repo.GetAll()
}

func (s *ConfigService) Create(conf *Config) (*Config, error) {
	return s.repo.Create(conf)
}

func (s *ConfigService) Update(conf *Config) (*Config, error) {
	return s.repo.Update(conf)
}

func (s *ConfigService) Delete(name string) error {
	return s.repo.Delete(name)
}

func (s *ConfigService) SetLastLoaded(id int) error {
	return s.repo.SetLastLoaded(id)
}

func (s *ConfigService) LastLoaded() (*Config, error) {
	return s.repo.LastLoaded()
}
