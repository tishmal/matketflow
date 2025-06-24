package models

import "time"

// PriceUpdate represents a price update from an exchange
type PriceUpdate struct {
	Exchange  string    `json:"exchange"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// ExchangeConfig holds configuration for each exchange
type ExchangeConfig struct {
	Name string
	Host string
	Port string
}
