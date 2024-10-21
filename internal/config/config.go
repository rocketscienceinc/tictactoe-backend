package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel   string `yaml:"log-level" env-default:"info"`
	HTTPPort   string `yaml:"http-port" env-default:"8080"`
	SocketPort string `yaml:"socket-port" env-default:"9090"`
	Redis      Redis  `yaml:"redis"`
}

type Redis struct {
	Host string `yaml:"host" env-default:"localhost"`
	Port string `yaml:"port" env-default:"6379"`
}

// MustLoad - load all configurations in config.yml file.
func MustLoad(path string) *Config {
	config := &Config{}

	if err := cleanenv.ReadConfig(path, config); err != nil {
		panic(fmt.Errorf("unable to load config file: %w", err))
	}

	return config
}

func (that *Redis) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", that.Host, that.Port)
}
