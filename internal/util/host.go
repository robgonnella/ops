package util

import (
	"os"
)

func Hostname() (*string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &hostname, nil
}
