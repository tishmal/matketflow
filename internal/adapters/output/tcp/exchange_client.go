package tcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"marketflow/internal/domain/models"
)

type TCPExchangeClient struct {
	logger *slog.Logger
	conn   net.Conn
}

// NEW METHOD
func NewTCPExchangeClient(logger *slog.Logger) *TCPExchangeClient {
	return &TCPExchangeClient{
		logger: logger,
	}
}

// ПЕРЕНЕСЕНО из connectAndListen
func (c *TCPExchangeClient) Connect(config models.ExchangeConfig) error {
	address := net.JoinHostPort(config.Host, config.Port)
	c.logger.Info("Connecting to exchange", "exchange", config.Name, "address", address)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", config.Name, err)
	}

	c.conn = conn
	c.logger.Info("Connected to exchange", "exchange", config.Name)
	return nil
}

// ПЕРЕНЕСЕНО и ИЗМЕНЕНО из connectAndListen
func (c *TCPExchangeClient) Listen(ctx context.Context, updates chan<- models.PriceUpdate) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	defer c.conn.Close()

	// Set read timeout
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	scanner := bufio.NewScanner(c.conn)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		update, err := c.parseMessage(line, "exchange") // TODO: передать имя биржи
		if err != nil {
			c.logger.Warn("Failed to parse message", "message", line, "error", err)
			continue
		}

		// Reset read timeout on successful read
		c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// Send to updates channel
		select {
		case updates <- update:
		case <-ctx.Done():
			return nil
		default:
			c.logger.Warn("Updates channel full, dropping update")
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return fmt.Errorf("connection closed")
}

func (c *TCPExchangeClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ПЕРЕНЕСЕНО из parseMessage (без изменений)
func (c *TCPExchangeClient) parseMessage(message, exchangeName string) (models.PriceUpdate, error) {
	// Try to parse as JSON first
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(message), &jsonData); err == nil {
		return c.parseJSONMessage(jsonData, exchangeName)
	}

	// Try to parse as simple format: SYMBOL:PRICE
	parts := strings.Split(message, ":")
	if len(parts) == 2 {
		symbol := strings.TrimSpace(parts[0])
		priceStr := strings.TrimSpace(parts[1])

		var price float64
		if _, err := fmt.Sscanf(priceStr, "%f", &price); err != nil {
			return models.PriceUpdate{}, fmt.Errorf("invalid price format: %s", priceStr)
		}

		return models.PriceUpdate{
			Exchange:  exchangeName,
			Symbol:    symbol,
			Price:     price,
			Timestamp: time.Now(),
		}, nil
	}

	// Try to parse as space-separated: SYMBOL PRICE
	parts = strings.Fields(message)
	if len(parts) >= 2 {
		symbol := parts[0]
		priceStr := parts[1]

		var price float64
		if _, err := fmt.Sscanf(priceStr, "%f", &price); err != nil {
			return models.PriceUpdate{}, fmt.Errorf("invalid price format: %s", priceStr)
		}

		return models.PriceUpdate{
			Exchange:  exchangeName,
			Symbol:    symbol,
			Price:     price,
			Timestamp: time.Now(),
		}, nil
	}

	return models.PriceUpdate{}, fmt.Errorf("unknown message format: %s", message)
}

// ПЕРЕНЕСЕНО из parseJSONMessage (без изменений)
func (c *TCPExchangeClient) parseJSONMessage(data map[string]interface{}, exchangeName string) (models.PriceUpdate, error) {
	symbol, ok := data["symbol"].(string)
	if !ok {
		if s, ok := data["pair"].(string); ok {
			symbol = s
		} else {
			return models.PriceUpdate{}, fmt.Errorf("missing symbol/pair field")
		}
	}

	var price float64
	if p, ok := data["price"].(float64); ok {
		price = p
	} else if p, ok := data["price"].(string); ok {
		if _, err := fmt.Sscanf(p, "%f", &price); err != nil {
			return models.PriceUpdate{}, fmt.Errorf("invalid price format: %s", p)
		}
	} else {
		return models.PriceUpdate{}, fmt.Errorf("missing or invalid price field")
	}

	return models.PriceUpdate{
		Exchange:  exchangeName,
		Symbol:    symbol,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}
