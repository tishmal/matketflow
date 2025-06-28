package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
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

			key := update.Exchange + ":" + update.Symbol
			score := float64(time.Now().Unix())

			// cохраняем все цены за минуту в Redis
			if err := s.redisClient.ZAdd(s.ctx, key, score, update.Price); err != nil {
				s.logger.Error("Failed to write to Redis ZSet", "error", err)
			}

			//получаем данные из Redis за минуту
			now := time.Now().Unix()
			min := fmt.Sprintf("%d", now-60)
			max := fmt.Sprintf("%d", now)

			values, err := s.redisClient.ZRangeByScore(s.ctx, key, min, max)
			if err != nil {
				s.logger.Error("Failed to get recent values from Redis", "error", err)
			}

			// парсим прайс
			var prices []float64
			for _, v := range values {
				price, err := strconv.ParseFloat(v, 64)
				if err != nil {
					s.logger.Error("Failed to parse price", "value", v, "error", err)
					continue
				}
				prices = append(prices, price)
			}

			// сохраняем из редиса в postgres
			err = s.db.InsertMarketData(update.Exchange, update.Symbol, prices, time.Now())
			if err != nil {
				s.logger.Error("Failed to save to PostgreSQL", "error", err)
			}

			// // Вывод данных из редиса
			// for _, val := range values {
			// 	if err := s.pricePublisher.PublishRedis(key, val, update); err != nil {
			// 		s.logger.Error("Failed to publish price update", "error", err)
			// 	}
			// }

			// удаляем из редиса устаревшие данные
			cutoff := time.Now().Add(-1 * time.Minute).Unix()
			err = s.redisClient.ZRemRangeByScore(s.ctx, key, "0", fmt.Sprintf("%d", cutoff))
			if err != nil {
				s.logger.Error("Failed to clean old Redis data", "error", err)
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
