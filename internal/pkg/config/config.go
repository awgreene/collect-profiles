package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	disabled = "disabled"
)

type config struct {
	Disabled bool
}

func GetConfig(path string) (*config, error) {
	file, err := os.Open(filepath.Join(path, "disabled"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadFile(filepath.Join(path, disabled))
	if err != nil {
		return nil, err
	}

	cfg := &config{}
	if strings.ToLower(string(data)) == "true" {
		cfg.Disabled = true
	}

	return cfg, nil
}
