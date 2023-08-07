package server

import (
	"errors"
	"net"
	"sync"

	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/exception"
	"github.com/robgonnella/ops/internal/logger"
)

// internal var for tracking event listeners
var channelID = 0

// generates sequential ids for registered event listeners
func nextChannelID() int {
	channelID++
	return channelID
}

// represents a registered event listener
type eventChannel struct {
	id   int
	send chan *event.Event
}

// helper for filtering registered event listeners
func filterChannels(channels []*eventChannel, fn func(c *eventChannel) bool) []*eventChannel {
	modifiedChannels := []*eventChannel{}
	for _, evtChan := range channels {
		if fn(evtChan) {
			modifiedChannels = append(modifiedChannels, evtChan)
		}
	}

	return modifiedChannels
}

// ServerService represents our server.Service implementation
type ServerService struct {
	log      logger.Logger
	repo     Repo
	evtChans []*eventChannel
	mux      sync.Mutex
}

// NewService returns a new instance ServerService
func NewService(conf config.Config, repo Repo) *ServerService {
	log := logger.New()

	return &ServerService{
		log:      log,
		repo:     repo,
		evtChans: []*eventChannel{},
		mux:      sync.Mutex{},
	}
}

// GetAllServers returns all servers from the database
func (s *ServerService) GetAllServers() ([]*Server, error) {
	return s.repo.GetAllServers()
}

// GetAllServersInNetworkTargets returns all servers in database that have ips
// within the provided list of network targets
func (s *ServerService) GetAllServersInNetworkTargets(targets []string) ([]*Server, error) {
	allServers, err := s.GetAllServers()

	result := []*Server{}

	if err != nil {
		return nil, err
	}

	for _, server := range allServers {
		for _, target := range targets {
			_, ipnet, err := net.ParseCIDR(target)

			if err != nil {
				// non CIDR target just check if target matches IP
				if server.IP == target {
					result = append(result, server)
					break
				}

				// target is not a cidr and does not match server ip
				// just continue looping targets
				continue
			}

			svrNetIP := net.ParseIP(server.IP)

			if ipnet != nil && ipnet.Contains(svrNetIP) {
				// server IP is within target CIDR block
				result = append(result, server)
				break
			}
		}
	}

	// s.log.Error().Interface("result", result).Msg("returning all servers in target")

	return result, nil
}

// AddOrUpdateServer adds or updates a server
func (s *ServerService) AddOrUpdateServer(req *Server) error {
	_, err := s.repo.GetServerByID(req.ID)

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

	updatedServer, err := s.repo.UpdateServer(req)

	if err != nil {
		return err
	}

	s.sendServerUpdateEvent(updatedServer)

	return nil
}

func (s *ServerService) UpdateMACByIP(req *Server) error {
	currentServer, err := s.repo.GetServerByIP(req.IP)

	if err != nil {
		return err
	}

	req.ID = currentServer.ID
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

// StreamEvents registers a listener for database updates
func (s *ServerService) StreamEvents(send chan *event.Event) int {
	evtChan := &eventChannel{
		id:   nextChannelID(),
		send: send,
	}

	s.mux.Lock()
	s.evtChans = append(s.evtChans, evtChan)
	s.mux.Unlock()

	return evtChan.id
}

// StopStream removes and closes channel for a specific registered listener
func (s *ServerService) StopStream(id int) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.log.Info().Int("channelID", id).Msg("Filtering channel")
	s.evtChans = filterChannels(s.evtChans, func(c *eventChannel) bool {
		if c.id == id {
			close(c.send)
		}
		return c.id != id
	})
}

// GetServer returns a single server from the database by ID if found
func (s *ServerService) GetServer(id string) (*Server, error) {
	return s.repo.GetServerByID(id)
}

// sends out server update events to all registered listeners
func (s *ServerService) sendServerUpdateEvent(server *Server) {
	s.mux.Lock()
	defer s.mux.Unlock()
	for _, clientChan := range s.evtChans {
		clientChan.send <- &event.Event{
			Type:    event.SeverUpdate,
			Payload: server,
		}
	}
}
