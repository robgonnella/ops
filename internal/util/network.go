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
	// udp doesn't make a full connection and will find the default ip
	// that traffic will use if say 2 are configured (wired and wireless)
	conn, err := net.Dial("udp", "8.8.8.8:80")

	var foundIP net.IP

	if err != nil {
		// resort to looping through interfaces
		ifaces, err := net.Interfaces()

		if err != nil {
			return nil, nil, err
		}

	OUTER:
		for _, i := range ifaces {
			addrs, err := i.Addrs()

			if err != nil {
				continue
			}

			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPAddr:
					if !v.IP.IsLoopback() {
						foundIP = v.IP
						break OUTER
					}
				}
			}
		}

		if foundIP == nil {
			return nil, nil, fmt.Errorf("failed to find IP address for this machine")
		}
	} else {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		foundIP = localAddr.IP
	}

	ipnet, err := getIPNetByIP(foundIP)

	if err != nil {
		return nil, nil, err
	}

	size, _ := ipnet.Mask.Size()

	ipCidr := fmt.Sprintf("%s/%d", foundIP, size)

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

	userIP := foundIP.String()
	cidr := fmt.Sprintf("%s/%d", firstCidrIP, size)

	return &userIP, &cidr, nil
}
