package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	EnvDBConnString = "DB_CONN_STRING"
)

type DB struct {
	ConnStr string `yaml:"conn_str,omitempty"`
}

type Server struct {
	Addr string `yaml:"addr,omitempty"`
}

type App struct {
	StartupTimeout              time.Duration `yaml:"startup_timeout"`
	SubscriptionsServiceTimeout time.Duration `yaml:"subscriptions_service_timeout,omitempty"`
}
type Config struct {
	DB     DB     `yaml:"db,omitempty"`
	Server Server `yaml:"server,omitempty"`
	App    App    `yaml:"app,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		DB: DB{ConnStr: "postgres://postgres:postgres@localhost:5527/postgres"},
		Server: Server{
			Addr: "127.0.0.1:8080",
		},
		App: App{SubscriptionsServiceTimeout: 10 * time.Second, StartupTimeout: 10 * time.Second},
	}
}

func ParseConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return cfg, err
	}
	if connStr := os.Getenv(EnvDBConnString); connStr != "" {
		cfg.DB.ConnStr = connStr
	}
	if err := file.Close(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func MustParseConfig(path string) Config {
	cfg, err := ParseConfig(path)
	if err != nil {
		panic(fmt.Sprintf("parse config: %s", err))
	}
	return cfg
}
