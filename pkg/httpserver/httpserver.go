package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/yokeTH/yoketh-backend-oss/pkg/apperror"
)

type Config struct {
	Env                  string `env:"ENV"`
	Name                 string `env:"NAME"`
	Port                 int    `env:"PORT"`
	BodyLimitMB          int    `env:"BODY_LIMIT_MB"`
	CorsAllowOrigins     string `env:"CORS_ALLOW_ORIGINS"`
	CorsAllowMethods     string `env:"CORS_ALLOW_METHODS"`
	CorsAllowHeaders     string `env:"CORS_ALLOW_HEADERS"`
	CorsAllowCredentials bool   `env:"CORS_ALLOW_CREDENTIALS"`
	SwaggerUser          string `env:"SWAGGER_USER"`
	SwaggerPass          string `env:"SWAGGER_PASS"`
}

const defaultEnv = "unknown"
const defaultName = "app"
const defaultPort = 8080
const defaultBodyLimitMB = 4
const defaultCorsAllowOrigins = "*"
const defaultCorsAllowMethods = "GET,POST,PUT,DELETE,PATCH,OPTIONS"
const defaultCorsAllowHeaders = "Origin,Content-Type,Accept,Authorization"
const defaultCorsAllowCredentials = false
const defalutSwaggerUser = "admin"
const defalutSwaggerPass = "1234"

type Server struct {
	config *Config
	*fiber.App
}

func New(opts ...ServerOption) *Server {

	defaultConfig := &Config{
		Env:                  defaultEnv,
		Name:                 defaultName,
		Port:                 defaultPort,
		BodyLimitMB:          defaultBodyLimitMB,
		CorsAllowOrigins:     defaultCorsAllowOrigins,
		CorsAllowMethods:     defaultCorsAllowMethods,
		CorsAllowHeaders:     defaultCorsAllowHeaders,
		CorsAllowCredentials: defaultCorsAllowCredentials,
		SwaggerUser:          defalutSwaggerUser,
		SwaggerPass:          defalutSwaggerPass,
	}

	server := &Server{
		config: defaultConfig,
	}

	for _, opt := range opts {
		opt(server)
	}

	app := fiber.New(fiber.Config{
		AppName:       server.config.Name,
		BodyLimit:     server.config.BodyLimitMB * 1024 * 1024,
		CaseSensitive: true,
		JSONEncoder:   json.Marshal,
		JSONDecoder:   json.Unmarshal,
		ErrorHandler:  apperror.ErrorHandler,
	})

	app.Use(requestid.New())

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	server.App = app

	return server
}

func (s *Server) Start(ctx context.Context, stop context.CancelFunc) {
	go func() {
		if err := s.Listen(fmt.Sprintf(":%d", s.config.Port)); err != nil {
			log.Fatalf("failed to start server: %v", err)
			stop()
		}
	}()

	defer func() {
		if err := s.Shutdown(); err != nil {
			log.Printf("failed to shutdown server: %v.", err)
		}
	}()

	<-ctx.Done()

	log.Println("shutting down server...")
}
