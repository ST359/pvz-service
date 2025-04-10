package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DbHost     string `env:"DATABASE_HOST"`
	DbPort     int    `env:"DATABASE_PORT"`
	DbUser     string `env:"DATABASE_USER"`
	DbPassword string `env:"DATABASE_PASSWORD"`
	DbName     string `env:"DATABASE_NAME"`
	Port       int    `env:"SERVER_PORT"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}
	return &cfg
}
