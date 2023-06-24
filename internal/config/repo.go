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
	return &SqliteRepo{
		db: db,
	}
}

// Get returns a config from the db
func (r *SqliteRepo) Get(id int) (*Config, error) {
	if id == 0 {
		return nil, errors.New("config id cannot be empty")
	}

	confModel := ConfigModel{ID: id}

	if result := r.db.First(&confModel); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrRecordNotFound
		}

		return nil, result.Error
	}

	if result := r.db.Save(&confModel); result.Error != nil {
		return nil, result.Error
	}

	return modelToConfig(&confModel)
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
		c, err := modelToConfig(&m)

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

	confModel, err := configToModel(conf)

	if err != nil {
		return nil, err
	}

	// create or update
	result := r.db.Create(confModel)

	if result.Error != nil {
		return nil, result.Error
	}

	return modelToConfig(confModel)
}

// Update updates a config in db
func (r *SqliteRepo) Update(conf *Config) (*Config, error) {
	if conf.ID == 0 {
		return nil, errors.New("config ID cannot be empty")
	}

	confModel, err := configToModel(conf)

	if err != nil {
		return nil, err
	}

	if result := r.db.Save(confModel); result.Error != nil {
		return nil, result.Error
	}

	return modelToConfig(confModel)
}

// Delete deletes a config from db
func (r *SqliteRepo) Delete(id int) error {
	if id == 0 {
		return errors.New("config id cannot be empty")
	}

	return r.db.Delete(&ConfigModel{ID: id}).Error
}

// SetLastLoaded updates a configs "loaded" field to the current timestamp
func (r *SqliteRepo) SetLastLoaded(id int) error {
	confModel := ConfigModel{ID: id}

	if result := r.db.First(&confModel); result.Error != nil {
		return result.Error
	}

	confModel.Loaded = time.Now()

	return r.db.Save(&confModel).Error
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

	return modelToConfig(&confModel)
}

// helpers
func modelToConfig(model *ConfigModel) (*Config, error) {
	overrides := []SSHOverride{}

	if err := json.Unmarshal([]byte(model.SSH.Overrides.String()), &overrides); err != nil {
		return nil, err
	}

	targets := []string{}

	if err := json.Unmarshal([]byte(model.Targets.String()), &targets); err != nil {
		return nil, err
	}

	return &Config{
		ID:   model.ID,
		Name: model.Name,
		SSH: SSHConfig{
			User:      model.SSH.User,
			Identity:  model.SSH.Identity,
			Overrides: overrides,
		},
		Targets: targets,
		Loaded:  model.Loaded,
	}, nil
}

func configToModel(conf *Config) (*ConfigModel, error) {
	overridesBytes, err := json.Marshal(conf.SSH.Overrides)

	if err != nil {
		return nil, err
	}

	targetsBytes, err := json.Marshal(conf.Targets)

	if err != nil {
		return nil, err
	}

	return &ConfigModel{
		ID:   conf.ID,
		Name: conf.Name,
		SSH: SSHConfigModel{
			User:      conf.SSH.User,
			Identity:  conf.SSH.Identity,
			Overrides: datatypes.JSON(overridesBytes),
		},
		Targets: datatypes.JSON(targetsBytes),
	}, nil
}
