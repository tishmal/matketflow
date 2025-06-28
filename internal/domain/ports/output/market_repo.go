package output

import "time"

type MarketRepository interface {
	InsertMarketData(exchange, symbol string, price float64, ts time.Time) error
}
