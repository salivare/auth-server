package config

import (
	"errors"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"os"
	"time"
)

type Config struct {
	Server ServerConfig `koanf:"server"`
	Token  TokenConfig  `koanf:"token"`
}

type ServerConfig struct {
	Addr         string        `koanf:"addr"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
	IdleTimeout  time.Duration `koanf:"idle_timeout"`
}

type TokenConfig struct {
	Secret     string        `koanf:"secret"`
	AccessTTL  time.Duration `koanf:"access_ttl"`
	RefreshTTL time.Duration `koanf:"refresh_ttl"`
}

func LoadConfig() (*Config, error) {
	patch := os.Getenv("CFG_APP_YAML_PATCH")
	if patch == "" {
		return nil, errors.New("CFG_APP_YAML_PATCH environment variable not set")
	}

	k := koanf.New(".")

	if err := k.Load(file.Provider(patch), yaml.Parser()); err != nil {
		return nil, err
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
