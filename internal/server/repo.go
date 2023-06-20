package server

import (
	"errors"

	"github.com/robgonnella/ops/internal/exception"
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

// GetAllServers returns all servers from the database
func (r *SqliteRepo) GetAllServers() ([]*Server, error) {
	servers := []*Server{}

	if result := r.db.Find(&servers); result.Error != nil {
		return nil, result.Error
	}

	return servers, nil
}

// GetServerByID returns a server from the SqliteRepo
func (r *SqliteRepo) GetServerByID(serverID string) (*Server, error) {
	server := Server{ID: serverID}

	if result := r.db.First(&server); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrRecordNotFound
		}

		return nil, result.Error
	}

	return &server, nil
}

// GetServerByIP returns a server from the SqliteRepo based on current ip
func (r *SqliteRepo) GetServerByIP(ip string) (*Server, error) {
	server := Server{IP: ip}

	if result := r.db.First(&server); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrRecordNotFound
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
