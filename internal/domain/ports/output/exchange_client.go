package output

import (
	"context"
	"marketflow/internal/domain/models"
)

type ExchangeClient interface {
	Connect(config models.ExchangeConfig) error
	Listen(ctx context.Context, updates chan<- models.PriceUpdate, exchange models.ExchangeConfig) error
	Close() error
}
