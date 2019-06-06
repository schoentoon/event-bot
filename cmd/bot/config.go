package main

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config structure of the config file
type Config struct {
	Telegram struct {
		Token string `yaml:"token"`
		Debug bool   `yaml:"debug"`
	} `yaml:"telegram"`
	Postgres struct {
		Addr string `yaml:"addr"`
	} `yaml:"postgres"`
	IDHash struct {
		Salt      string `yaml:"salt"`
		MinLength int    `yaml:"minLength"`
	} `yaml:"idhash"`
	Workers   int    `yaml:"workers"`
	Templates string `yaml:"templates"`
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

	err = checkConfig(out)
	if err != nil {
		return nil, err
	}

	return out, err
}

func checkConfig(cfg *Config) error {
	if cfg.Workers <= 0 {
		return fmt.Errorf("%d workers configured, needs to be more than 0", cfg.Workers)
	}

	if cfg.Templates == "" {
		return errors.New("No templates directory configured")
	}

	if cfg.IDHash.Salt == "" {
		return errors.New("No salt configured!")
	}

	return nil
}
