package config

import "time"

type DB struct {
	ConnStr string `yaml:"conn_str"`
}

type Server struct {
	Addr string `yaml:"addr"`
}

type App struct {
	SubscriptionsServiceTimeout time.Duration `yaml:"timeouts"`
}
type Config struct {
	DB     DB     `yaml:"db"`
	Server Server `yaml:"server"`
	App    App    `yaml:"app"`
}

func DefaultConfig() Config {
	return Config{
		DB: DB{ConnStr: "postgres://postgres:postgres@localhost:5527/postgres"},
		Server: Server{
			Addr: "127.0.0.1:8080",
		},
		App: App{SubscriptionsServiceTimeout: 10 * time.Second},
	}
}
