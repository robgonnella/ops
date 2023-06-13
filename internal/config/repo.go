package config

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/robgonnella/opi/internal/exception"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SqliteRepo is our repo implementation for sqlite
type SqliteRepo struct {
	db *gorm.DB
}

// NewSqliteDatabase returns a new opi sqlite db
func NewSqliteDatabase(db *gorm.DB) *SqliteRepo {
	return &SqliteRepo{db: db}
}

// Get returns a config from the db
func (r *SqliteRepo) Get(name string) (*Config, error) {
	if name == "" {
		return nil, errors.New("config name cannot be empty")
	}

	confModel := ConfigModel{Name: name}

	if result := r.db.First(&confModel); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrRecordNotFound
		}

		return nil, result.Error
	}

	confModel.Loaded = time.Now()

	if result := r.db.Save(&confModel); result.Error != nil {
		return nil, result.Error
	}

	conf := &Config{}

	if err := json.Unmarshal([]byte(confModel.Data.String()), conf); err != nil {
		return nil, err
	}

	return conf, nil
}

// Create creates a new config in db
func (r *SqliteRepo) Create(conf *Config) (*Config, error) {
	if conf.Name == "" {
		return nil, errors.New("config name cannot be empty")
	}

	dataBytes, err := json.Marshal(conf)

	if err != nil {
		return nil, err
	}

	confModel := &ConfigModel{
		Name: conf.Name,
		Data: datatypes.JSON(dataBytes),
	}

	if result := r.db.Create(confModel); result.Error != nil {
		return nil, result.Error
	}

	return r.Get(conf.Name)
}

// Update updates a config in db
func (r *SqliteRepo) Update(conf *Config) (*Config, error) {
	if conf.Name == "" {
		return nil, errors.New("config name cannot be empty")
	}

	dataBytes, err := json.Marshal(conf)

	if err != nil {
		return nil, err
	}

	confModel := &ConfigModel{
		Name: conf.Name,
		Data: datatypes.JSON(dataBytes),
	}

	if result := r.db.Save(confModel); result.Error != nil {
		return nil, result.Error
	}

	return r.Get(conf.Name)
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

	conf := &Config{}

	if err := json.Unmarshal([]byte(confModel.Data.String()), conf); err != nil {
		return nil, err
	}

	return conf, nil
}
