package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"marketflow/internal/adapters/input/cli"
	"marketflow/internal/adapters/output/console"
	"marketflow/internal/adapters/output/tcp"
	"marketflow/internal/config"
	"marketflow/internal/domain/services"

	pgx "github.com/jackc/pgx/v5"
	redis "github.com/redis/go-redis/v9"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("Load configuration failed", "error", err)
		os.Exit(1)
	}

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addrRedis := fmt.Sprintf(cfg.Redis.Host + ":" + cfg.Redis.Port)

	// redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     addrRedis,          // адрес Redis-сервера
		Password: cfg.Redis.Password, // если нет пароля
		DB:       cfg.Redis.DB,       // номер БД
	})

	// postgres
	connString := fmt.Sprintf("%s://%s:%s@%s:%s/%s", cfg.Postgres.Host, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.NameDB)
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// инициализация репо...

	// Create output adapters
	exchangeClient := tcp.NewTCPExchangeClient(logger)
	pricePublisher := console.NewConsolePricePublisher(logger)

	// Create domain service
	marketService := services.NewMarketService(
		ctx,
		exchangeClient,
		pricePublisher,
		cfg.Exchanges,
		logger,
		rdb,
		cfg.RedisTTL,
	)

	// Create input adapter
	cliHandler := cli.NewCLIHandler(ctx, marketService, logger)

	// Start application
	if err := cliHandler.Start(); err != nil {
		logger.Error("Application failed", "error", err)
		os.Exit(1)
	}
}
