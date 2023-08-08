package discovery

import (
	"bytes"
	"context"
	"net"
	"regexp"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/projectdiscovery/mapcidr"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/server"
	"github.com/robgonnella/ops/internal/util"
	"github.com/rs/zerolog/log"
)

var cidrSuffix = regexp.MustCompile(`\/\d{2}$`)

type ARPScanner struct {
	ctx          context.Context
	cancel       context.CancelFunc
	targets      []string
	networkInfo  *util.NetworkInfo
	handle       *pcap.Handle
	packetSource *gopacket.PacketSource
	arpMap       map[string]net.HardwareAddr
	resultChan   chan *DiscoveryResult
	mux          sync.RWMutex
	log          logger.Logger
}

func NewARPScanner(
	networkInfo *util.NetworkInfo,
	targets []string,
	resultChan chan *DiscoveryResult,
) (*ARPScanner, error) {
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

	// Open up a pcap handle for packet reads/writes.
	handle, err := pcap.OpenLive(
		networkInfo.Interface.Name,
		65536,
		true,
		pcap.BlockForever,
	)

	if err != nil {
		return nil, err
	}

	src := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)

	ctx, cancel := context.WithCancel(context.Background())

	scanner := &ARPScanner{
		ctx:          ctx,
		cancel:       cancel,
		targets:      ipList,
		handle:       handle,
		packetSource: src,
		networkInfo:  networkInfo,
		arpMap:       map[string]net.HardwareAddr{},
		resultChan:   resultChan,
		mux:          sync.RWMutex{},
		log:          logger.New(),
	}

	go scanner.readPackets()

	return scanner, nil
}

func (s *ARPScanner) Scan() error {
	return s.writeARP()
}

func (s *ARPScanner) Stop() {
	s.cancel()
	s.handle.Close()
}

func (s *ARPScanner) readPackets() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case packet := <-s.packetSource.Packets():
			arpLayer := packet.Layer(layers.LayerTypeARP)

			if arpLayer != nil {
				s.handleARPLayer(arpLayer.(*layers.ARP))
				continue
			}

			s.handleNonARPPacket(packet)
		}
	}
}

func (s *ARPScanner) handleARPLayer(arp *layers.ARP) {
	if arp.Operation != layers.ARPReply {
		// not an arp reply
		return
	}

	if bytes.Equal([]byte(s.networkInfo.Interface.HardwareAddr), arp.SourceHwAddress) {
		// This is a packet we sent
		return
	}

	ip := net.IP(arp.SourceProtAddress)
	mac := net.HardwareAddr(arp.SourceHwAddress)

	if !util.SliceIncludes(s.targets, ip.String()) {
		// not an arp request we care about
		return
	}

	s.mux.Lock()
	s.arpMap[ip.String()] = mac
	s.mux.Unlock()

	s.writeSyn(ip, mac)
}

func (s *ARPScanner) handleNonARPPacket(packet gopacket.Packet) {
	net := packet.NetworkLayer()

	if net == nil {
		return
	}

	srcIP := net.NetworkFlow().Src().String()

	s.mux.RLock()
	mac, ok := s.arpMap[srcIP]
	s.mux.RUnlock()

	if !ok {
		return
	}

	tcpLayer := packet.Layer(layers.LayerTypeTCP)

	if tcpLayer == nil {
		return
	}

	tcp := tcpLayer.(*layers.TCP)

	if tcp.DstPort != 54321 {
		return
	}

	if tcp.SrcPort != 22 {
		return
	}

	res := &DiscoveryResult{
		ID:       mac.String(),
		IP:       srcIP,
		Hostname: "",
		OS:       "",
		Status:   server.StatusOnline,
	}

	port := Port{
		ID:     22,
		Status: PortClosed,
	}

	if tcp.SYN && tcp.ACK {
		port.Status = PortOpen
	}

	res.Ports = []Port{port}

	s.resultChan <- res
}

func (s *ARPScanner) writeARP() error {
	eth := layers.Ethernet{
		SrcMAC:       s.networkInfo.Interface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}

	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(s.networkInfo.Interface.HardwareAddr),
		SourceProtAddress: []byte(s.networkInfo.UserIP.To4()),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
	}

	buf := gopacket.NewSerializeBuffer()

	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	for _, ip := range s.targets {
		arp.DstProtAddress = []byte(net.ParseIP(ip).To4())

		if err := gopacket.SerializeLayers(buf, opts, &eth, &arp); err != nil {
			return err
		}

		if err := s.handle.WritePacketData(buf.Bytes()); err != nil {
			log.Error().Err(err).Msg("")
		}
	}

	return nil
}

func (s *ARPScanner) writeSyn(ip net.IP, mac net.HardwareAddr) {
	eth := layers.Ethernet{
		SrcMAC:       s.networkInfo.Interface.HardwareAddr,
		DstMAC:       mac,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ip4 := layers.IPv4{
		SrcIP:    s.networkInfo.UserIP.To4(),
		DstIP:    ip.To4(),
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}

	tcp := layers.TCP{
		SrcPort: 54321,
		DstPort: 22,
		SYN:     true,
	}

	tcp.SetNetworkLayerForChecksum(&ip4)

	buf := gopacket.NewSerializeBuffer()

	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	if err := gopacket.SerializeLayers(buf, opts, &eth, &ip4, &tcp); err != nil {
		s.log.Error().Err(err).Msg("failed to serialize syn layers")
		return
	}

	if err := s.handle.WritePacketData(buf.Bytes()); err != nil {
		s.log.Error().Err(err).Msg("failed to send syn packet")
	}
}
