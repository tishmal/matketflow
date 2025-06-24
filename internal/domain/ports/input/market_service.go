package input

import "context"

type MarketService interface {
	Start(ctx context.Context) error
	Stop() error
}
