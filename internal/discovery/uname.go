package discovery

import (
	"context"
	"os/exec"
	"regexp"
	"strings"

	"github.com/robgonnella/ops/internal/config"
)

var osReleaseRegexp = regexp.MustCompile(`PRETTY_NAME=(?P<os>.+)`)

// UnameScanner is an implementation of the DetailScanner interface
type UnameScanner struct {
	conf config.Config
}

// NewUnameScanner returns a new instance of UnameScanner
func NewUnameScanner(conf config.Config) *UnameScanner {
	return &UnameScanner{conf: conf}
}

// GetServerDetails returns server details using ssh and "uname -a" command
func (s UnameScanner) GetServerDetails(ctx context.Context, ip string) (*Details, error) {
	user := s.conf.SSH.User
	identity := s.conf.SSH.Identity
	port := s.conf.SSH.Port

	for _, o := range s.conf.SSH.Overrides {
		if o.Target == ip {
			if o.User != "" {
				user = o.User
			}

			if o.Identity != "" {
				identity = o.Identity
			}

			if o.Port != "" {
				port = o.Port
			}
		}
	}

	unameCmd := exec.Command(
		"ssh",
		"-i",
		identity,
		"-p",
		port,
		"-o",
		"BatchMode=yes",
		"-o",
		"StrictHostKeyChecking=no",
		"-l",
		user,
		ip,
		"uname -a",
	)

	unameOutput, err := unameCmd.Output()

	if err != nil {
		return nil, err
	}

	info := strings.Split(string(unameOutput), " ")

	operatingSystem := info[0]
	hostname := info[1]

	switch operatingSystem {
	case "Darwin":
		operatingSystem = "MacOS"
	case "Linux":
		osReleaseCmd := exec.Command(
			"ssh",
			"-i",
			identity,
			"-p",
			port,
			"-o",
			"BatchMode=yes",
			"-o",
			"StrictHostKeyChecking=no",
			"-l",
			user,
			ip,
			"cat /etc/os-release",
		)

		osReleaseOutput, err := osReleaseCmd.Output()

		if err != nil {
			return nil, err
		}

		match := osReleaseRegexp.FindStringSubmatch(string(osReleaseOutput))

		for i, name := range osReleaseRegexp.SubexpNames() {
			if name == "os" {
				operatingSystem = match[i]
			}
		}
	}

	return &Details{
		Hostname: hostname,
		OS:       operatingSystem,
	}, nil
}
