package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"marketflow/internal/domain/models"
	"marketflow/internal/domain/ports/output"
)

type MarketServiceImpl struct {
	exchanges      []models.ExchangeConfig
	exchangeClient output.ExchangeClient
	pricePublisher output.PricePublisher
	dataChan       chan models.PriceUpdate
	wg             sync.WaitGroup
	ctx            context.Context
	logger         *slog.Logger
	redisTTL       time.Duration
	redisClient    output.RedisClient
	db             output.MarketRepository
	reconnectCh    chan models.ExchangeConfig
	knownKeys      map[string]struct{}
	mu             sync.RWMutex
}

// NEW METHOD - заменяет NewMarketDataProcessor
func NewMarketService(
	ctx context.Context,
	exchangeClient output.ExchangeClient,
	pricePublisher output.PricePublisher,
	exchanges []models.ExchangeConfig,
	logger *slog.Logger,
	redisClient output.RedisClient,
	db output.MarketRepository,
	redisTTL time.Duration,
) *MarketServiceImpl {
	return &MarketServiceImpl{
		exchanges:      exchanges,
		exchangeClient: exchangeClient,
		pricePublisher: pricePublisher,
		dataChan:       make(chan models.PriceUpdate, 1000),
		ctx:            ctx,
		logger:         logger,
		redisClient:    redisClient,
		db:             db,
		redisTTL:       redisTTL,
		reconnectCh:    make(chan models.ExchangeConfig, 10),
		knownKeys:      make(map[string]struct{}),
	}
}

// ПЕРЕНЕСЕННЫЕ МЕТОДЫ (изменены для работы с интерфейсами)
func (s *MarketServiceImpl) Start(ctx context.Context) error {
	s.logger.Info("Starting MarketFlow Live Mode")

	// Start data collector (Fan-In pattern)
	go s.dataCollector()

	go s.aggregator()

	// Start reconnection handler
	go s.reconnectionHandler()

	// Start exchange listeners (Fan-Out pattern)
	for _, exchange := range s.exchanges {
		s.wg.Add(1)
		go s.listenToExchange(exchange)
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for all goroutines
	s.wg.Wait()
	return nil
}

func (s *MarketServiceImpl) Stop() error {
	s.logger.Info("Stopping MarketFlow")
	close(s.dataChan)
	close(s.reconnectCh)
	return nil
}

func (s *MarketServiceImpl) dataCollector() {
	s.logger.Info("Starting data collector")

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Data collector stopped")
			return

		case update, ok := <-s.dataChan:
			if !ok {
				s.logger.Info("Data channel closed")
				return
			}

			key := update.Exchange + ":" + update.Pair
			score := float64(time.Now().Unix())

			// Сохраняем цену в Redis (ZSet)
			if err := s.redisClient.ZAdd(s.ctx, key, score, update.Price); err != nil {
				s.logger.Error("Failed to write to Redis ZSet", "error", err)
			}

			// Добавляем ключ в список известных для агрегатора
			s.mu.Lock()
			s.knownKeys[key] = struct{}{}
			s.mu.Unlock()

			// Удаляем устаревшие данные старше 60 секунд
			cutoff := time.Now().Add(-1 * time.Minute).Unix()
			if err := s.redisClient.ZRemRangeByScore(s.ctx, key, "0", fmt.Sprintf("%d", cutoff)); err != nil {
				s.logger.Error("Failed to clean old Redis data", "error", err)
			}
		}
	}
}

func (s *MarketServiceImpl) aggregator() {
	s.logger.Info("Aggregator started")
	ticker := time.NewTicker(10 * time.Second) // временно для отладки
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Aggregator stopped")
			return
		case <-ticker.C:
			s.mu.RLock()
			keys := make([]string, 0, len(s.knownKeys))
			for k := range s.knownKeys {
				keys = append(keys, k)
			}
			s.mu.RUnlock()

			now := time.Now().Unix()
			min := fmt.Sprintf("%d", now-60)
			max := fmt.Sprintf("%d", now)

			for _, key := range keys {
				parts := strings.Split(key, ":")
				if len(parts) != 2 {
					continue
				}
				ex := parts[0]
				pair := parts[1]

				values, err := s.redisClient.ZRangeByScore(s.ctx, key, min, max)
				if err != nil {
					s.logger.Error("Aggregator Redis read error", "key", key, "error", err)
					continue
				}

				var prices []float64
				for _, v := range values {
					price, err := strconv.ParseFloat(v, 64)
					if err != nil {
						s.logger.Error("Parse error", "value", v, "error", err)
						continue
					}
					prices = append(prices, price)
				}

				if len(prices) == 0 {
					s.logger.Info("No prices to write", "key", key)
					continue
				}

				err = s.db.InsertMarketData(ex, pair, prices, time.Now())
				if err != nil {
					s.logger.Error("DB insert failed", "error", err)
				} else {
					s.logger.Info("Wrote to DB", "exchange", ex, "pair", pair, "count", len(prices))
				}
			}
		}
	}
}

// ИЗМЕНЕНО: теперь использует ExchangeClient интерфейс
func (s *MarketServiceImpl) listenToExchange(exchange models.ExchangeConfig) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Exchange listener stopped", "exchange", exchange.Name)
			return
		default:
			if err := s.exchangeClient.Connect(exchange); err != nil {
				s.logger.Error("Connection failed", "exchange", exchange.Name, "error", err)

				// Schedule reconnection
				select {
				case s.reconnectCh <- exchange:
				case <-s.ctx.Done():
					return
				}

				// Wait before retry
				select {
				case <-time.After(5 * time.Second):
				case <-s.ctx.Done():
					return
				}
			} else {
				// Listen for updates
				if err := s.exchangeClient.Listen(s.ctx, s.dataChan, exchange); err != nil {
					s.logger.Error("Listen failed", "exchange", exchange.Name, "error", err)
				}
			}
		}
	}
}

func (s *MarketServiceImpl) reconnectionHandler() {
	s.logger.Info("Starting reconnection handler")

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Reconnection handler stopped")
			return
		case exchange := <-s.reconnectCh:
			s.logger.Info("Scheduling reconnection", "exchange", exchange.Name)
		}
	}
}
