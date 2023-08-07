package discovery

import (
	"context"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/projectdiscovery/mapcidr"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/server"
	"github.com/robgonnella/ops/internal/util"
)

type PCapScanner struct {
	ctx          context.Context
	cancel       context.CancelFunc
	networkInfo  *util.NetworkInfo
	targets      []string
	handle       *pcap.Handle
	packetSource *gopacket.PacketSource
	log          logger.Logger
}

func NewPCapScanner(
	targets []string,
	networkInfo *util.NetworkInfo,
) (*PCapScanner, error) {
	ipList := []string{}

	for _, t := range targets {
		if cidrSuffix.MatchString(t) {
			ips, err := mapcidr.IPAddresses(t)

			if err != nil {
				return nil, err
			}

			ipList = append(ipList, ips...)
		} else {
			ipList = append(ipList, t)
		}
	}

	handle, err := pcap.OpenLive(
		networkInfo.Interface.Name,
		int32(networkInfo.Interface.MTU),
		true,
		pcap.BlockForever,
	)

	if err != nil {
		return nil, err
	}

	packetSource := gopacket.NewPacketSource(handle, layers.LinkTypeEthernet)

	ctx, cancel := context.WithCancel(context.Background())

	return &PCapScanner{
		ctx:          ctx,
		cancel:       cancel,
		networkInfo:  networkInfo,
		targets:      ipList,
		handle:       handle,
		packetSource: packetSource,
		log:          logger.New(),
	}, nil
}

func (s *PCapScanner) Stop() {
	s.cancel()
	s.handle.Close()
}

func (s *PCapScanner) ListenForPackets(res chan *server.Server) {
	for packet := range s.packetSource.Packets() {
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		ethLayer := packet.Layer(layers.LayerTypeEthernet)

		if ipLayer == nil || ethLayer == nil {
			continue
		}

		ipv4 := ipLayer.(*layers.IPv4)
		eth := ethLayer.(*layers.Ethernet)

		if ipv4.SrcIP.Equal(s.networkInfo.UserIP) {
			continue
		}

		srcIP := ipv4.SrcIP.String()
		srcMAC := eth.SrcMAC.String()

		if util.SliceIncludes(s.targets, srcIP) {
			res <- &server.Server{IP: srcIP, MAC: srcMAC}
		}
	}
}
