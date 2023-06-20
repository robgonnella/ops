package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strings"

	"github.com/robgonnella/ops/cli/commands"
	"github.com/robgonnella/ops/internal/logger"
	"github.com/robgonnella/ops/internal/name"
	"github.com/robgonnella/ops/internal/ui"
	"github.com/spf13/viper"
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

// get userIP and cidr block for preferred outbound ip of this machine
func getNetworkInfo() (*string, *string, error) {
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

func setRuntTimeConfig() error {
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	configDir := path.Join(userHomeDir, ".config", name.APP_NAME)

	if err := os.Mkdir(configDir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	logFile := path.Join(configDir, name.APP_NAME+".log")

	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		return err
	}

	cacheDir := path.Join(userCacheDir, name.APP_NAME)

	if err := os.Mkdir(cacheDir, 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	dbFile := path.Join(cacheDir, name.APP_NAME+".db")

	userIP, cidr, err := getNetworkInfo()

	if err != nil {
		return err
	}

	defaultSSHIdentity := path.Join(userHomeDir, ".ssh", "id_rsa")

	user := os.Getenv("USER")

	// share run-time config globally using viper
	viper.Set("log-file", logFile)
	viper.Set("config-dir", configDir)
	viper.Set("cache-dir", cacheDir)
	viper.Set("database-file", dbFile)
	viper.Set("default-cidr", *cidr)
	viper.Set("user-ip", *userIP)
	viper.Set("default-ssh-identity", defaultSSHIdentity)
	viper.Set("user", user)

	return nil
}

// Entry point for the cli.
func main() {
	log := logger.New()

	err := setRuntTimeConfig()

	if err != nil {
		log.Fatal().Err(err)
	}

	appUI := ui.New()

	// Get the "root" cobra cli command
	cmd := commands.Root(&commands.CommandProps{
		UI: appUI,
	})

	// execute the cobra command and exit with error code if necessary
	err = cmd.ExecuteContext(context.Background())

	if err != nil {
		log.Fatal().Err(err)
	}
}
