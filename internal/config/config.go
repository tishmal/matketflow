package config

import (
	"marketflow/internal/domain/models"
)

type Config struct {
	Exchanges []models.ExchangeConfig
}

// NEW METHOD
func NewConfig() *Config {
	return &Config{
		Exchanges: []models.ExchangeConfig{
			{Name: "Exchange1", Host: "exchange1", Port: "40101"},
			{Name: "Exchange2", Host: "exchange2", Port: "40102"},
			{Name: "Exchange3", Host: "exchange3", Port: "40103"},
		},
	}
}

// // ПЕРЕНЕСЕНО из getExchangeHost (без изменений)
// func getExchangeHost(exchangeName string) string {
// 	// Check if running in Docker container
// 	if host := os.Getenv(strings.ToUpper(exchangeName) + "_HOST"); host != "" {
// 		return host
// 	}
// 	// Default to localhost for local development
// 	return "127.0.0.1"
// }
