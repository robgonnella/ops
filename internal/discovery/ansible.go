package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/apenella/go-ansible/pkg/adhoc"
	"github.com/apenella/go-ansible/pkg/execute"
	"github.com/apenella/go-ansible/pkg/execute/measure"
	"github.com/apenella/go-ansible/pkg/options"
	"github.com/robgonnella/opi/internal/config"
)

type AnsibleIpScanner struct {
	conf config.Config
}

func NewAnsibleIpScanner(conf config.Config) *AnsibleIpScanner {
	return &AnsibleIpScanner{conf: conf}
}

func (s *AnsibleIpScanner) GetServerDetails(ctx context.Context, ip string) (*AnsibleDetails, error) {
	if err := os.Setenv(options.AnsibleHostKeyCheckingEnv, "False"); err != nil {
		return nil, err
	}

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

	buff := new(bytes.Buffer)

	executorTimeMeasurement := measure.NewExecutorTimeMeasurement(
		execute.NewDefaultExecute(
			execute.WithWrite(io.Writer(buff)),
		),
	)

	ansibleConnectionOptions := &options.AnsibleConnectionOptions{
		Connection: "ssh",
		User:       user,
		PrivateKey: identity,
	}

	ansibleAdhocOptions := &adhoc.AnsibleAdhocOptions{
		ModuleName: "setup",
		Inventory:  ip + ",",
	}

	adhoc := &adhoc.AnsibleAdhocCmd{
		Pattern:           "all",
		Options:           ansibleAdhocOptions,
		ConnectionOptions: ansibleConnectionOptions,
		Exec:              executorTimeMeasurement,
		StdoutCallback:    "json",
	}

	if err := adhoc.Run(ctx); err != nil {
		return nil, err
	}

	str := strings.Replace(buff.String(), ip+" | SUCCESS =>", "", 1)

	details := map[string]interface{}{}

	if err := json.Unmarshal([]byte(str), &details); err != nil {
		return nil, err
	}

	facts := details["ansible_facts"].(map[string]interface{})
	hostname := facts["ansible_hostname"].(string)
	os := facts["ansible_distribution"].(string)

	return &AnsibleDetails{
		Hostname: hostname,
		OS:       os,
	}, nil
}
