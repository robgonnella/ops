package util

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/jackpal/gateway"
)

type NetworkInfo struct {
	Hostname  string
	Interface *net.Interface
	Gateway   net.IP
	UserIP    net.IP
	Cidr      string
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func hostname() (*string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &hostname, nil
}

// get network interface associated with ip
func getIPNetByIP(ip net.IP) (*net.Interface, *net.IPNet, error) {
	interfaces, err := net.Interfaces()

	if err != nil {
		return nil, nil, err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()

		if err != nil {
			continue
		}

		for _, addr := range addrs {
			_, ipnet, err := net.ParseCIDR(addr.String())

			if err != nil {
				continue
			}

			if ipnet.Contains(ip) {
				return &iface, ipnet, nil
			}
		}
	}

	return nil, nil, errors.New("failed to find IPNet")
}

// GetNetworkInfo returns userIP and cidr block for preferred
// outbound ip of this machine
func GetNetworkInfo() (*NetworkInfo, error) {
	gw, err := gateway.DiscoverGateway()

	if err != nil {
		return nil, err
	}

	host, err := hostname()

	if err != nil {
		return nil, err
	}

	// udp doesn't make a full connection and will find the default ip
	// that traffic will use if say 2 are configured (wired and wireless)
	conn, err := net.Dial("udp", gw.String()+":80")

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	foundIP := net.ParseIP(localAddr.IP.String())

	iface, ipnet, err := getIPNetByIP(foundIP)

	if err != nil {
		return nil, err
	}

	size, _ := ipnet.Mask.Size()

	ipCidr := fmt.Sprintf("%s/%d", foundIP.String(), size)

	ip, ipnet, err := net.ParseCIDR(ipCidr)

	if err != nil {
		return nil, err
	}

	firstCidrIP := ""

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		if !strings.HasSuffix(ip.String(), ".0") {
			firstCidrIP = ip.String()
			break
		}
	}

	cidr := fmt.Sprintf("%s/%d", firstCidrIP, size)

	return &NetworkInfo{
		Hostname:  *host,
		Interface: iface,
		Gateway:   gw,
		UserIP:    foundIP,
		Cidr:      cidr,
	}, nil
}
