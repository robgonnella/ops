package config

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/robgonnella/ops/internal/exception"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SqliteRepo is our repo implementation for sqlite
type SqliteRepo struct {
	db *gorm.DB
}

// NewSqliteRepo returns a new ops sqlite db
func NewSqliteRepo(db *gorm.DB) *SqliteRepo {
	return &SqliteRepo{db: db}
}

// Get returns a config from the db
func (r *SqliteRepo) Get(name string) (*Config, error) {
	if name == "" {
		return nil, errors.New("config name cannot be empty")
	}

	confModel := ConfigModel{}

	if result := r.db.First(&confModel, "name = ?", name); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrRecordNotFound
		}

		return nil, result.Error
	}

	confModel.Loaded = time.Now()

	if result := r.db.Save(&confModel); result.Error != nil {
		return nil, result.Error
	}

	overrides := []SSHOverride{}

	if err := json.Unmarshal([]byte(confModel.SSH.Overrides.String()), &overrides); err != nil {
		return nil, err
	}

	targets := []string{}

	if err := json.Unmarshal([]byte(confModel.Targets.String()), &targets); err != nil {
		return nil, err
	}

	conf := &Config{
		ID:   confModel.ID,
		Name: confModel.Name,
		SSH: SSHConfig{
			User:      confModel.SSH.User,
			Identity:  confModel.SSH.Identity,
			Overrides: overrides,
		},
		Targets: targets,
		Loaded:  confModel.Loaded,
	}

	return conf, nil
}

// GetAll returns all configs in db
func (r *SqliteRepo) GetAll() ([]*Config, error) {
	confModels := []ConfigModel{}

	if result := r.db.Find(&confModels); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrRecordNotFound
		}

		return nil, result.Error
	}

	confs := []*Config{}

	for _, m := range confModels {
		c, err := r.Get(m.Name)

		if err != nil {
			return nil, err
		}

		confs = append(confs, c)
	}

	return confs, nil
}

// Create creates a new config in db
func (r *SqliteRepo) Create(conf *Config) (*Config, error) {
	if conf.Name == "" {
		return nil, errors.New("config name cannot be empty")
	}

	overridesBytes, err := json.Marshal(conf.SSH.Overrides)

	if err != nil {
		return nil, err
	}

	targetsBytes, err := json.Marshal(conf.Targets)

	if err != nil {
		return nil, err
	}

	confModel := &ConfigModel{
		Name: conf.Name,
		SSH: SSHConfigModel{
			User:      conf.SSH.User,
			Identity:  conf.SSH.Identity,
			Overrides: datatypes.JSON(overridesBytes),
		},
		Targets: datatypes.JSON(targetsBytes),
	}

	// create or update
	result := r.db.
		Where(&ConfigModel{Name: confModel.Name}).
		Assign(confModel).
		FirstOrCreate(confModel)

	if result.Error != nil {
		return nil, result.Error
	}

	return r.Get(conf.Name)
}

// Update updates a config in db
func (r *SqliteRepo) Update(conf *Config) (*Config, error) {
	if conf.ID == 0 {
		return nil, errors.New("config ID cannot be empty")
	}

	foundModel := ConfigModel{ID: conf.ID}

	if result := r.db.First(&foundModel); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrRecordNotFound
		}

		return nil, result.Error
	}

	overridesByes, err := json.Marshal(conf.SSH.Overrides)

	if err != nil {
		return nil, err
	}

	targetBytes, err := json.Marshal(conf.Targets)

	if err != nil {
		return nil, err
	}

	confModel := &ConfigModel{
		ID:   foundModel.ID,
		Name: conf.Name,
		SSH: SSHConfigModel{
			User:      conf.SSH.User,
			Identity:  conf.SSH.Identity,
			Overrides: datatypes.JSON(overridesByes),
		},
		Targets: datatypes.JSON(targetBytes),
	}

	if result := r.db.Save(confModel); result.Error != nil {
		return nil, result.Error
	}

	return r.Get(conf.Name)
}

// Delete deletes a config from db
func (r *SqliteRepo) Delete(name string) error {
	if name == "" {
		return errors.New("config name cannot be empty")
	}

	return r.db.Where("name = ?", name).Delete(&ConfigModel{Name: name}).Error
}

// LastLoaded returns the most recently loaded config
func (r *SqliteRepo) LastLoaded() (*Config, error) {
	confModel := ConfigModel{}

	if result := r.db.Order("loaded desc").First(&confModel); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrRecordNotFound
		}

		return nil, result.Error
	}

	return r.Get(confModel.Name)
}
