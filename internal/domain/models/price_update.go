package models

import "time"

type PriceUpdate struct {
	Exchange  string    `json:"exchange"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

type ExchangeConfig struct {
	Name string
	Host string
	Port string
}
