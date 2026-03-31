package config

import (
	"log"
	"log/slog"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"API_ADDRESS" env-default:"localhost:80"`
	Timeout time.Duration `yaml:"timeout" env:"API_TIMEOUT" env-default:"5s"`
}

type Config struct {
	LogLevel      string     `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	HTTPConfig    HTTPConfig `yaml:"api_server"`
	WordsAddress  string     `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"words:81"`
	UpdateAddress string     `yaml:"update_address" env:"UPDATE_ADDRESS" env-default:"update:82"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		slog.Error("error reading server config file:", "error", err)
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		slog.Error("error reading server env:", "error", err)
		log.Fatalf("cannot read env : %s", err)
	}

	return cfg
}
