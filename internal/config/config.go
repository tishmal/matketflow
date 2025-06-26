package config

import (
	"fmt"
	"log"
	"marketflow/internal/domain/models"
	"marketflow/pkg/utils"
	"os"
	"path/filepath"
	"time"
)

// func NewConfig() *Config {
// 	return &Config{
// 		Exchanges: []models.ExchangeConfig{
// 			{Name: "Exchange1", Host: "exchange1", Port: "40101"},
// 			{Name: "Exchange2", Host: "exchange2", Port: "40102"},
// 			{Name: "Exchange3", Host: "exchange3", Port: "40103"},
// 		},
// 	}
// }

type Config struct {
	Postgres         PostgresConfig
	Redis            RedisConfig
	Exchanges        []models.ExchangeConfig
	PortAPI          int
	AggregatorWindow time.Duration
	RedisTTL         time.Duration
	AppEnv           string
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	NameDB   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Name     string
	Password string
	DB       int
}

func NewConfig() (*Config, error) {
	if err := utils.LoadEnv(filepath.Join(".env")); err != nil {
		log.Fatalf("Ошибка загрузки .env: %v", err)
	}

	env := map[string]string{
		"PG_HOST":           os.Getenv("PG_HOST"),
		"PG_PORT":           os.Getenv("PG_PORT"),
		"PG_USER":           os.Getenv("PG_USER"),
		"PG_PASSWORD":       os.Getenv("PG_PASSWORD"),
		"PG_NAME":           os.Getenv("PG_NAME"),
		"PG_SSLMODE":        os.Getenv("PG_SSLMODE"),
		"REDIS_HOST":        os.Getenv("REDIS_HOST"),
		"REDIS_PORT":        os.Getenv("REDIS_PORT"),
		"REDIS_PASSWORD":    os.Getenv("REDIS_PASSWORD"),
		"REDIS_DB":          os.Getenv("REDIS_DB"),
		"EXCHANGE1_NAME":    os.Getenv("EXCHANGE1_NAME"),
		"EXCHANGE2_NAME":    os.Getenv("EXCHANGE2_NAME"),
		"EXCHANGE3_NAME":    os.Getenv("EXCHANGE3_NAME"),
		"EXCHANGE1_PORT":    os.Getenv("EXCHANGE1_PORT"),
		"EXCHANGE2_PORT":    os.Getenv("EXCHANGE2_PORT"),
		"EXCHANGE3_PORT":    os.Getenv("EXCHANGE3_PORT"),
		"REDIS_TLS":         os.Getenv("REDIS_TLS"),
		"AGGREGATOR_WINDOW": os.Getenv("AGGREGATOR_WINDOW"),
	}
	log.Println(os.Getenv("EXCHANGE1_NAME"))
	for key, value := range env {
		if value == "" {
			return nil, fmt.Errorf("missing required env variable: %s", key)
		}
	}
	pgPort, err := utils.ParseEnvInt("PG_PORT")
	if err != nil {
		return nil, err
	}

	redisPort, err := utils.ParseEnvInt("REDIS_PORT")
	if err != nil {
		return nil, err
	}

	redisDB, err := utils.ParseEnvInt("REDIS_DB")
	if err != nil {
		return nil, err
	}

	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("EXCHANGE%d_PORT", i)
		_, err := utils.ParseEnvInt(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", key, err)
		}
	}

	portAPI, err := utils.ParseEnvInt("API_PORT")
	if err != nil {
		return nil, err
	}

	aggregatorWindow, err := utils.ValidTime("AGGREGATOR_WINDOW")
	if err != nil {
		return nil, err
	}

	redisTTL, err := utils.ValidTime("REDIS_TTL")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Postgres: PostgresConfig{
			Host:     os.Getenv("PG_HOST"),
			Port:     pgPort,
			User:     os.Getenv("PG_USER"),
			Password: os.Getenv("PG_PASSWORD"),
			NameDB:   os.Getenv("POSTGRES_DB"),
			SSLMode:  os.Getenv("PG_SSLMODE"),
		},
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     redisPort,
			Name:     os.Getenv("REDIS_DB"),
			Password: os.Getenv("PG_PASSWORD"),
			DB:       redisDB,
		},
		Exchanges: []models.ExchangeConfig{
			{
				Name: os.Getenv("EXCHANGE1_NAME"),
				Host: os.Getenv("EXCHANGE1_NAME"),
				Port: os.Getenv("EXCHANGE1_PORT"),
			},
			{
				Name: os.Getenv("EXCHANGE2_NAME"),
				Host: os.Getenv("EXCHANGE2_NAME"),
				Port: os.Getenv("EXCHANGE2_PORT"),
			},
			{
				Name: os.Getenv("EXCHANGE3_NAME"),
				Host: os.Getenv("EXCHANGE3_NAME"),
				Port: os.Getenv("EXCHANGE3_PORT"),
			},
		},
		PortAPI:          portAPI,
		AggregatorWindow: aggregatorWindow,
		RedisTTL:         redisTTL,
		AppEnv:           os.Getenv("APP_ENV"),
	}

	return cfg, nil
}
