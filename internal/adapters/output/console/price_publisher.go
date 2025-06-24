package console

import (
	"fmt"
	"log/slog"

	"marketflow/internal/domain/models"
)

type ConsolePricePublisher struct {
	logger *slog.Logger
}

// NEW METHOD
func NewConsolePricePublisher(logger *slog.Logger) *ConsolePricePublisher {
	return &ConsolePricePublisher{
		logger: logger,
	}
}

// ПЕРЕНЕСЕНО из printPriceUpdate
func (p *ConsolePricePublisher) Publish(update models.PriceUpdate) error {
	fmt.Printf("[%s] %s - %s: $%.6f\n",
		update.Timestamp.Format("15:04:05.000"),
		update.Exchange,
		update.Symbol,
		update.Price)
	return nil
}
