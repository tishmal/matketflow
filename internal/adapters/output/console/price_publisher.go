package console

import (
	"fmt"
	"log/slog"
	"strconv"

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

func (p *ConsolePricePublisher) PublishRedis(key, value string, update models.PriceUpdate) error {
	fValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		p.logger.Error("Ошибка:", err)
		return err
	}

	fmt.Printf("[%s] %s: $%.6f\n",
		update.Timestamp.Format("15:04:05.000"),
		key,
		fValue)
	return nil
}

// func (p *ConsolePricePublisher) Publish(update models.PriceUpdate) error {
// 	fmt.Printf("[%s] %s - %s: $%.6f\n",
// 		update.Timestamp.Format("15:04:05.000"),
// 		update.Exchange,
// 		update.Symbol,
// 		update.Price)
// 	return nil
// }
