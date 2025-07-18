package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"marketflow/internal/domain/ports/input"
)

type CLIHandler struct {
	marketService input.MarketService
	logger        *slog.Logger
	ctx           context.Context
}

// NEW METHOD
func NewCLIHandler(ctx context.Context, marketService input.MarketService, logger *slog.Logger) *CLIHandler {
	return &CLIHandler{
		marketService: marketService,
		logger:        logger,
		ctx:           ctx,
	}
}

// логика обработки сигналов
func (h *CLIHandler) Start() error {
	// Handle command line arguments
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Usage:")
		fmt.Println("  marketflow [--port <N>]")
		fmt.Println("  marketflow --help")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  --port N     Port number")
		return nil
	}

	//

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, stopping gracefully...")
		h.marketService.Stop()
	}()

	// Start processing
	return h.marketService.Start(h.ctx)
}
