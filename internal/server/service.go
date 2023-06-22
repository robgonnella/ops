package server

import (
	"context"
	"errors"
	"net"

	"github.com/imdario/mergo"
	"github.com/robgonnella/ops/internal/config"
	"github.com/robgonnella/ops/internal/event"
	"github.com/robgonnella/ops/internal/exception"
	"github.com/robgonnella/ops/internal/logger"
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
	log      logger.Logger
	repo     Repo
	evtChans []*eventChannel
}

// NewService returns a new instance server service
func NewService(conf config.Config, repo Repo) *ServerService {
	log := logger.New()

	ctx := context.Background()

	return &ServerService{
		ctx:      ctx,
		log:      log,
		repo:     repo,
		evtChans: []*eventChannel{},
	}
}

// GetAllServers returns all servers from the database
func (s *ServerService) GetAllServers() ([]*Server, error) {
	return s.repo.GetAllServers()
}

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
					s.log.Debug().
						Str("serverIP", server.IP).
						Str("target", target).
						Msg("serverIP matches network target")
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
				s.log.Debug().
					Str("serverIP", server.IP).
					Str("target", target).
					Msg("network target cidr includes serverIP")
				result = append(result, server)
				break
			}
		}
	}

	return result, nil
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
	s.log.Info().Int("channelID", id).Msg("Filtering channel")
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
