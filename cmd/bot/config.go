package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config structure of the config file
type Config struct {
	Telegram struct {
		Token  string `yaml:"token"`
		Debug  bool   `yaml:"debug"`
		Buffer int    `yaml:"buffer"`
	} `yaml:"telegram"`
	Postgres struct {
		Addr               string        `yaml:"addr"`
		MaxOpenConnections int           `yaml:"maxOpenConns"`
		MaxIdleConnections int           `yaml:"maxIdleConns"`
		MaxConnLifetime    time.Duration `yaml:"maxConnLifetime"`
	} `yaml:"postgres"`
	IDHash struct {
		Salt      string `yaml:"salt"`
		MinLength int    `yaml:"minLength"`
	} `yaml:"idhash"`
	Prometheus struct {
		ListenAddr string `yaml:"addr"`
	} `yaml:"prometheus"`
	Sentry struct {
		Enabled          bool   `yaml:"enabled"`
		Dsn              string `yaml:"dsn"`
		AttachStacktrace bool   `yaml:"stacktrace"`
		Release          string `yaml:"release"`
	} `yaml:"sentry"`
	Workers              int           `yaml:"workers"`
	Templates            string        `yaml:"templates"`
	EventRefreshInterval time.Duration `yaml:"eventRefreshInterval"`
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

	if cfg.EventRefreshInterval == 0 {
		return errors.New("No event refresh interval set")
	}

	if cfg.Telegram.Buffer == 0 {
		return errors.New("No Telegram buffer size set")
	}

	return nil
}

func (c *Config) ApplyDatabase(db *sql.DB) {
	if c.Postgres.MaxConnLifetime > 0 {
		db.SetConnMaxLifetime(c.Postgres.MaxConnLifetime)
	}
	if c.Postgres.MaxIdleConnections > 0 {
		db.SetMaxIdleConns(c.Postgres.MaxIdleConnections)
	}
	if c.Postgres.MaxOpenConnections > 0 {
		db.SetMaxOpenConns(c.Postgres.MaxOpenConnections)
	}
}
