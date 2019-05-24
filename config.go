package main

import (
	"os"
	"gopkg.in/yaml.v2"
)

// Config structure of the config file
type Config struct {
	Telegram struct {
		Token string `yaml:"token"`
		Debug bool `yaml:"debug"`
	} `yaml:"telegram"`
	Postgres struct {
		Addr string `yaml:"addr"`
	} `yaml:"postgres"`
}

// ReadConfig reads a file into the config structure
func ReadConfig(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	out := &Config{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, err
}