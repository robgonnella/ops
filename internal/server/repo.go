package server

import (
	"errors"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ErrNotFound custom database error for failure to find record
var ErrNotFound = errors.New("record not found")

// SqliteRepo is our repo implementation for sqlite
type SqliteRepo struct {
	db *gorm.DB
}

// NewSqliteDatabase returns a new opi sqlite db
func NewSqliteDatabase() (*SqliteRepo, error) {
	filepath := viper.Get("database-file")

	dbFile, ok := filepath.(string)

	if !ok {
		return nil, errors.New("failed to find database file path config")
	}

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Server{})

	return &SqliteRepo{db: db}, nil
}

// GetAllServers returns all servers from the database
func (r *SqliteRepo) GetAllServers() ([]*Server, error) {
	servers := []*Server{}

	if result := r.db.Find(&servers); result.Error != nil {
		return nil, result.Error
	}

	return servers, nil
}

// GetServerByID returns a server from the SqliteDatabase
func (r *SqliteRepo) GetServerByID(serverID string) (*Server, error) {
	server := Server{ID: serverID}

	if result := r.db.First(&server); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, result.Error
	}

	return &server, nil
}

// GetServerByIP returns a server from the SqliteDatabase based on current ip
func (r *SqliteRepo) GetServerByIP(ip string) (*Server, error) {
	server := Server{IP: ip}

	if result := r.db.First(&server); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, result.Error
	}

	return &server, nil
}

func (r *SqliteRepo) AddServer(server *Server) (*Server, error) {
	if server.ID == "" {
		return nil, errors.New("server id cannot be empty")
	}

	if result := r.db.Create(server); result.Error != nil {
		return nil, result.Error
	}

	return server, nil
}

func (r *SqliteRepo) RemoveServer(id string) error {
	if id == "" {
		return errors.New("server id cannot be empty")
	}

	return r.db.Delete(&Server{ID: id}).Error
}

func (r *SqliteRepo) UpdateServer(server *Server) (*Server, error) {
	if server.ID == "" {
		return nil, errors.New("server id cannot be empty")
	}

	if result := r.db.Save(server); result.Error != nil {
		return nil, result.Error
	}

	return server, nil
}
