package output

import "time"

type MarketRepository interface {
	InsertMarketData(exchange, symbol string, prices []float64, ts time.Time) error
}
