package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"environment"`
	LogLevel   string `yaml:"log-level" env-default:"info"`
	HTTPPort   string `yaml:"http-port" env-default:"8080"`
	SocketPort string `yaml:"socket-port" env-default:"9090"`
}

// MustLoad - load all configurations in config.yml file.
func MustLoad(path string) *Config {
	config := &Config{}

	if err := cleanenv.ReadConfig(path, config); err != nil {
		panic(fmt.Errorf("unable to load config file: %w", err))
	}

	return config
}
