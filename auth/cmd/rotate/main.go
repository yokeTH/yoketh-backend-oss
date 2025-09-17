package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3/log"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/yokeTH/yoketh-backend-oss/auth/internal/config"
	"github.com/yokeTH/yoketh-backend-oss/auth/internal/db"
	"github.com/yokeTH/yoketh-backend-oss/auth/internal/key"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Warnf("Unable to load .env file: %s", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg := config.NewFromEnv()
	dbCfg := cfg.DBConfig

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.DBName, dbCfg.SSLMode)
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	queries := db.New(sqlDB)

	keyWrapper, err := key.NewLocalWrapperFromEnv()
	if err != nil {
		panic(err)
	}
	keyManager := key.NewManager(sqlDB, queries, keyWrapper, "auth-service")
	if err := keyManager.Init(ctx); err != nil {
		panic(err)
	}

	keyManager.Rotate(ctx)

	stop()
}
