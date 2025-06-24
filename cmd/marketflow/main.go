package main

import (
	"log/slog"
	"os"

	"marketflow/internal/adapters/input/cli"
	"marketflow/internal/adapters/output/console"
	"marketflow/internal/adapters/output/tcp"
	"marketflow/internal/config"
	"marketflow/internal/domain/services"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg := config.NewConfig()

	// Create output adapters
	exchangeClient := tcp.NewTCPExchangeClient(logger)
	pricePublisher := console.NewConsolePricePublisher(logger)

	// Create domain service
	marketService := services.NewMarketService(
		exchangeClient,
		pricePublisher,
		cfg.Exchanges,
		logger,
	)

	// Create input adapter
	cliHandler := cli.NewCLIHandler(marketService, logger)

	// Start application
	if err := cliHandler.Start(); err != nil {
		logger.Error("Application failed", "error", err)
		os.Exit(1)
	}
}
