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

	for _, o := range s.conf.SSH.Overrides {
		if o.Target == ip {
			if o.User != "" {
				user = o.User
			}

			if o.Identity != "" {
				identity = o.Identity
			}
		}
	}

	cmd := exec.Command("ssh", "-i", identity, user+"@"+ip, "uname -a")

	unameOutput, err := cmd.Output()

	if err != nil {
		return nil, err
	}

	info := strings.Split(string(unameOutput), " ")

	os := info[0]
	hostname := info[1]

	switch os {
	case "Darwin":
		os = "MacOS"
	case "Linux":
		cmd = exec.Command("ssh", "-i", identity, user+"@"+ip, "cat /etc/os-release")

		osReleaseOutput, err := cmd.Output()

		if err != nil {
			return nil, err
		}

		match := osReleaseRegexp.FindStringSubmatch(string(osReleaseOutput))

		for i, name := range osReleaseRegexp.SubexpNames() {
			if name == "os" {
				os = match[i]
			}
		}
	}

	return &Details{
		Hostname: hostname,
		OS:       os,
	}, nil
}
