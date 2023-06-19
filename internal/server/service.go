package server

import (
	"context"
	"errors"

	"github.com/imdario/mergo"
	"github.com/robgonnella/opi/internal/config"
	"github.com/robgonnella/opi/internal/event"
	"github.com/robgonnella/opi/internal/exception"
	"github.com/robgonnella/opi/internal/logger"
)

var channelID = 0

func nextChannelID() int {
	channelID++
	return channelID
}

type eventChannel struct {
	id   int
	send chan *event.Event
}

func filterChannels(channels []*eventChannel, fn func(c *eventChannel) bool) []*eventChannel {
	modifiedChannels := []*eventChannel{}
	for _, evtChan := range channels {
		if fn(evtChan) {
			modifiedChannels = append(modifiedChannels, evtChan)
		}
	}

	return modifiedChannels
}

// ServerService represents our server service implementation
type ServerService struct {
	ctx      context.Context
	logger   logger.Logger
	repo     Repo
	evtChans []*eventChannel
}

// NewService returns a new instance server service
func NewService(conf config.Config, repo Repo) *ServerService {
	log := logger.New()

	ctx := context.Background()

	return &ServerService{
		ctx:      ctx,
		logger:   log,
		repo:     repo,
		evtChans: []*eventChannel{},
	}
}

// GetAllServers returns all servers from the database
func (s *ServerService) GetAllServers() ([]*Server, error) {
	return s.repo.GetAllServers()
}

// AddOrUpdateServer adds or updates a server
func (s *ServerService) AddOrUpdateServer(req *Server) error {
	currentServer, err := s.repo.GetServerByID(req.ID)

	if errors.Is(err, exception.ErrRecordNotFound) {
		// handle add case
		updatedServer, err2 := s.repo.AddServer(req)

		if err2 != nil {
			return err2
		}

		s.sendServerUpdateEvent(updatedServer)

		return nil
	}

	if err != nil {
		// handle all other errors
		return err
	}

	// handle update case

	mergo.Merge(req, currentServer)

	updatedServer, err := s.repo.UpdateServer(req)

	if err != nil {
		return err
	}

	s.sendServerUpdateEvent(updatedServer)

	return nil
}

// MarkDeviceOffline marks a device offline based on it's current IP address
func (s *ServerService) MarkServerOffline(ip string) error {
	server, err := s.repo.GetServerByIP(ip)

	if errors.Is(err, exception.ErrRecordNotFound) {
		// no server found - don't return error here as there's no need to
		// mark a non-existent server as offline
		return nil
	}

	if err != nil {
		// handle all other errors
		return err
	}

	server.Status = StatusOffline
	server.SshStatus = SSHDisabled

	_, err = s.repo.UpdateServer(server)

	return err
}

// StreamEvents streams server updates to client
func (s *ServerService) StreamEvents(send chan *event.Event) int {
	evtChan := &eventChannel{
		id:   nextChannelID(),
		send: send,
	}

	s.evtChans = append(s.evtChans, evtChan)

	return evtChan.id
}

func (s *ServerService) StopStream(id int) {
	s.logger.Info().Int("channelID", id).Msg("Filtering channel")
	s.evtChans = filterChannels(s.evtChans, func(c *eventChannel) bool {
		return c.id != id
	})
}

// GetServer returns a single server from the database by ID if found
func (s *ServerService) GetServer(id string) (*Server, error) {
	return s.repo.GetServerByID(id)
}

func (s *ServerService) sendServerUpdateEvent(server *Server) {
	for _, clientChan := range s.evtChans {
		clientChan.send <- &event.Event{
			Type:    event.SeverUpdate,
			Payload: server,
		}
	}
}
