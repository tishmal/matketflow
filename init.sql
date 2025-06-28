-- Создание базы данных marketflow (если не существует)
-- CREATE DATABASE marketflow;

-- Подключаемся к базе данных marketflow
-- \c marketflow;

-- Создание таблицы для хранения агрегированных данных
CREATE TABLE IF NOT EXISTS market_data (
    id SERIAL PRIMARY KEY,
    pair_name VARCHAR(20) NOT NULL,
    exchange VARCHAR(20) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    average_price DECIMAL(20,8) NOT NULL,
    min_price DECIMAL(20,8),
    max_price DECIMAL(20,8),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_market_data_pair_exchange ON market_data(pair_name, exchange);
CREATE INDEX IF NOT EXISTS idx_market_data_timestamp ON market_data(timestamp);
CREATE INDEX IF NOT EXISTS idx_market_data_pair_timestamp ON market_data(pair_name, timestamp);
CREATE INDEX IF NOT EXISTS idx_market_data_exchange_timestamp ON market_data(exchange, timestamp);

-- Создание таблицы для хранения сырых данных (опционально, для debugging)
CREATE TABLE IF NOT EXISTS raw_price_data (
    id SERIAL PRIMARY KEY,
    pair_name VARCHAR(20) NOT NULL,
    exchange VARCHAR(20) NOT NULL,
    price DECIMAL(20,8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для таблицы сырых данных
CREATE INDEX IF NOT EXISTS idx_raw_price_data_pair_exchange ON raw_price_data(pair_name, exchange);
CREATE INDEX IF NOT EXISTS idx_raw_price_data_timestamp ON raw_price_data(timestamp);
CREATE INDEX IF NOT EXISTS idx_raw_price_data_received_at ON raw_price_data(received_at);

-- Создание функции для очистки старых данных (старше 30 дней)
CREATE OR REPLACE FUNCTION cleanup_old_data()
RETURNS void AS $$
BEGIN
    -- Удаляем данные старше 30 дней из market_data
    DELETE FROM market_data 
    WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '30 days';
    
    -- Удаляем данные старше 7 дней из raw_price_data
    DELETE FROM raw_price_data 
    WHERE received_at < CURRENT_TIMESTAMP - INTERVAL '7 days';
    
    -- Логируем количество удаленных записей
    RAISE NOTICE 'Cleanup completed at %', CURRENT_TIMESTAMP;
END;
$$ LANGUAGE plpgsql;

-- Создание представления для получения последних цен по парам
CREATE OR REPLACE VIEW latest_prices AS
SELECT DISTINCT ON (pair_name, exchange)
    pair_name,
    exchange,
    average_price as price,
    timestamp
FROM market_data
ORDER BY pair_name, exchange, timestamp DESC;

-- Создание представления для получения статистики по парам за последний час
CREATE OR REPLACE VIEW hourly_stats AS
SELECT 
    pair_name,
    exchange,
    AVG(average_price) as avg_price,
    MIN(min_price) as min_price,
    MAX(max_price) as max_price,
    COUNT(*) as data_points,
    MIN(timestamp) as period_start,
    MAX(timestamp) as period_end
FROM market_data 
WHERE timestamp >= CURRENT_TIMESTAMP - INTERVAL '1 hour'
GROUP BY pair_name, exchange;

-- Вставка примерных данных для тестирования (опционально)
-- INSERT INTO market_data (pair_name, exchange, timestamp, average_price, min_price, max_price)
-- VALUES 
--     ('BTCUSDT', 'exchange1', CURRENT_TIMESTAMP - INTERVAL '1 minute', 45000.00, 44950.00, 45050.00),
--     ('ETHUSDT', 'exchange1', CURRENT_TIMESTAMP - INTERVAL '1 minute', 3000.00, 2995.00, 3005.00),
--     ('BTCUSDT', 'exchange2', CURRENT_TIMESTAMP - INTERVAL '1 minute', 45010.00, 44960.00, 45060.00);

-- Создание пользователя для приложения (если необходимо)
-- CREATE USER marketflow_app WITH PASSWORD 'secure_password';
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO marketflow_app;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO marketflow_app;

COMMIT;