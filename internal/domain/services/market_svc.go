package services

import (
	"context"
	"log/slog"
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
	cancel         context.CancelFunc
	logger         *slog.Logger
	reconnectCh    chan models.ExchangeConfig
}

// NEW METHOD - заменяет NewMarketDataProcessor
func NewMarketService(
	exchangeClient output.ExchangeClient,
	pricePublisher output.PricePublisher,
	exchanges []models.ExchangeConfig,
	logger *slog.Logger,
) *MarketServiceImpl {
	ctx, cancel := context.WithCancel(context.Background())

	return &MarketServiceImpl{
		exchanges:      exchanges,
		exchangeClient: exchangeClient,
		pricePublisher: pricePublisher,
		dataChan:       make(chan models.PriceUpdate, 1000),
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger,
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
	s.cancel()
	close(s.dataChan)
	close(s.reconnectCh)
	return nil
}

// ИЗМЕНЕНО: теперь использует PricePublisher интерфейс
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
			// Используем интерфейс вместо прямого вывода
			if err := s.pricePublisher.Publish(update); err != nil {
				s.logger.Error("Failed to publish price update", "error", err)
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
				if err := s.exchangeClient.Listen(s.ctx, s.dataChan); err != nil {
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
