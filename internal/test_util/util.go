package test_util

import (
	"github.com/robgonnella/go-lanscan/pkg/network"
)

func GetTestInterfaceName() (string, error) {
	netInfo, err := network.NewDefaultNetwork()

	if err != nil {
		return "", err
	}

	return netInfo.Interface().Name, nil
}
