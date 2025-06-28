package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"marketflow/internal/domain/models"
	"marketflow/internal/domain/ports/output"

	redis "github.com/redis/go-redis/v9"
)

type MarketServiceImpl struct {
	exchanges      []models.ExchangeConfig
	exchangeClient output.ExchangeClient
	pricePublisher output.PricePublisher
	dataChan       chan models.PriceUpdate
	wg             sync.WaitGroup
	ctx            context.Context
	logger         *slog.Logger
	redisClient    *redis.Client
	redisTTL       time.Duration
	reconnectCh    chan models.ExchangeConfig
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
		redisTTL:       redisTTL,
		reconnectCh:    make(chan models.ExchangeConfig, 10),
	}
}

// ПЕРЕНЕСЕННЫЕ МЕТОДЫ (изменены для работы с интерфейсами)
func (s *MarketServiceImpl) Start(ctx context.Context) error {
	s.logger.Info("Starting MarketFlow Live Mode")

	// Start data collector (Fan-In pattern)
	go s.dataCollector()

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
				return
			}
			// Установка значения в Redis
			key := update.Exchange + ":" + update.Symbol
			if err := s.redisClient.Set(s.ctx, key, update.Price, s.redisTTL).Err(); err != nil {
				s.logger.Error("Failed to write to Redis", "error", err)
			}

			// Получение значения
			val, err := s.redisClient.Get(s.ctx, key).Result()
			if err != nil {
				panic(err)
			}

			// Используем интерфейс вместо прямого вывода
			if err := s.pricePublisher.PublishRedis(key, val, update); err != nil {
				s.logger.Error("Failed to publish price update", "error", err)
			}

			// сохраняем в postgres
			err = s.db.InsertMarketData(update.Exchange, update.Symbol, update.Price, time.Now())
			if err != nil {
				s.logger.Error("Failed to save to PostgreSQL", "error", err)
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
