package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type MarketRepo struct {
	conn *pgx.Conn
	ctx  context.Context
}

func NewMarketRepo(ctx context.Context, conn *pgx.Conn) *MarketRepo {
	return &MarketRepo{conn: conn, ctx: ctx}
}

func (r *MarketRepo) InsertMarketData(exchange, symbol string, price float64, ts time.Time) error {
	_, err := r.conn.Exec(
		context.Background(),
		`INSERT INTO market_data (exchange, pair_name, average_price, timestamp) VALUES ($1, $2, $3, $4)`,
		exchange, symbol, price, ts,
	)
	return err
}
