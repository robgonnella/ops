package util

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// get network interface associated with ip
func getIPNetByIP(ip net.IP) (*net.IPNet, error) {
	interfaces, err := net.Interfaces()

	if err != nil {
		return nil, err
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
				return ipnet, nil
			}
		}
	}

	return nil, errors.New("failed to find IPNet")
}

// GetNetworkInfo returns userIP and cidr block for preferred
// outbound ip of this machine
func GetNetworkInfo() (*string, *string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		return nil, nil, err
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	ipnet, err := getIPNetByIP(localAddr.IP)

	if err != nil {
		return nil, nil, err
	}

	size, _ := ipnet.Mask.Size()

	ipCidr := fmt.Sprintf("%s/%d", localAddr.IP, size)

	ip, ipnet, err := net.ParseCIDR(ipCidr)
	if err != nil {
		return nil, nil, err
	}

	firstCidrIP := ""

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		if !strings.HasSuffix(ip.String(), ".0") {
			firstCidrIP = ip.String()
			break
		}
	}

	userIP := localAddr.IP.String()
	cidr := fmt.Sprintf("%s/%d", firstCidrIP, size)

	return &userIP, &cidr, nil
}
