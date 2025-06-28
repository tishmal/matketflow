package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type MarketRepo struct {
	conn *pgx.Conn
	ctx  context.Context
	log  *slog.Logger
}

func NewMarketRepo(ctx context.Context, conn *pgx.Conn, log *slog.Logger) *MarketRepo {
	return &MarketRepo{conn: conn, ctx: ctx, log: log}
}

func (r *MarketRepo) InsertMarketData(exchange, symbol string, prices []float64, ts time.Time) error {
	avg, min, max, ok := calculateStats(prices)
	if !ok {
		r.log.Warn("No prices to calculate statistics")
		return fmt.Errorf("no prices to calculate statistics")
	}

	_, err := r.conn.Exec(
		context.Background(),
		`INSERT INTO market_data (exchange, pair_name, average_price, min_price, max_price, timestamp) VALUES ($1, $2, $3, $4, $5, $6)`,
		exchange, symbol, avg, min, max, ts,
	)
	return err
}

func calculateStats(prices []float64) (avg, min, max float64, ok bool) {
	if len(prices) == 0 {
		return 0, 0, 0, false
	}

	min = prices[0]
	max = prices[0]
	sum := 0.0

	for _, p := range prices {
		if p < min {
			min = p
		}
		if p > max {
			max = p
		}
		sum += p
	}

	avg = sum / float64(len(prices))
	return avg, min, max, true
}
