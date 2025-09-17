package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/gofiber/fiber/v3/log"
	"github.com/yokeTH/yoketh-backend-oss/auth/internal/db"
	"github.com/yokeTH/yoketh-backend-oss/pkg/httpserver"
)

type config struct {
	Server   httpserver.Config `envPrefix:"SERVER_"`
	DBConfig db.Config         `envPrefix:"DATABASE_"`
}

func NewFromEnv() *config {
	config := &config{}

	if err := env.Parse(config); err != nil {
		log.Fatalf("Unable to parse env vars: %s", err)
	}

	return config
}
