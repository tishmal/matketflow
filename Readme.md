# MarketFlow - Real-Time Market Data Processor

MarketFlow - это система обработки рыночных данных в реальном времени, построенная на Go с использованием паттернов конкурентного программирования.

## 🚀 Быстрый старт

### Предварительные требования

- Docker и Docker Compose
- Go 1.24.2+ (для локальной разработки)
- `netcat` для тестирования соединений

### 1. Подготовка файлов

Убедитесь, что у вас есть следующие файлы в директории проекта:

**Файлы exchange (скачайте нужные для вашей архитектуры):**
- `exchange1_amd64.tar` или `exchange1_arm64.tar`
- `exchange2_amd64.tar` или `exchange2_arm64.tar` 
- `exchange3_amd64.tar` или `exchange3_arm64.tar`

### 2. Определение архитектуры

```bash
# Проверьте архитектуру вашей системы
uname -m

# x86_64 = amd64
# arm64/aarch64 = arm64
```

### 3. 🏁 Запуск (локально)

#### Ручная настройка
```bash
# 1. Загрузить Docker образы exchanges
docker load -i exchange1_arm64.tar  # или amd64
docker load -i exchange2_arm64.tar
docker load -i exchange3_arm64.tar

# 2. Обновить docker-compose.yml для вашей архитектуры
# (замените arm64 на amd64 если нужно)

# 3. Запустить систему
docker-compose up --build -d
```

## 📊 Использование

### Просмотр данных в реальном времени

```bash
# Логи MarketFlow
docker-compose logs -f marketflow

# Пример вывода:
# [14:23:15.123] Exchange1 - BTCUSDT: $43250.500000
# [14:23:15.234] Exchange2 - ETHUSDT: $2680.750000
# [14:23:15.345] Exchange3 - SOLUSDT: $98.250000
```

## 🔧 Конфигурация

### Структура проекта

```
marketflow/
├── cmd/
│   └── marketflow/
│       └── main.go                 # Точка входа приложения
├── internal/
│   ├── domain/                     # Доменный слой (центр гексагона)
│   │   ├── models/
│   │   │   └── price_update.go     # PriceUpdate struct
│   │   ├── ports/
│   │   │   ├── input/
│   │   │   │   └── market_service.go    # MarketService interface
│   │   │   └── output/
│   │   │       ├── exchange_client.go   # ExchangeClient interface
│   │   │       └── price_publisher.go   # PricePublisher interface
│   │   └── services/
│   │       └── market_service.go   # MarketServiceImpl
│   ├── adapters/                   # Адаптеры (края гексагона)
│   │   ├── input/
│   │   │   └── cli/
│   │   │       └── handler.go      # CLI обработчик
│   │   └── output/
│   │       ├── tcp/
│   │       │   └── exchange_client.go  # TCP клиент для бирж
│   │       └── console/
│   │           └── price_publisher.go  # Консольный вывод цен
│   └── config/
│       └── config.go               # Конфигурация приложения
├── pkg/
│   └── logger/
│       └── logger.go               # Настройка логгера
├── go.mod
├── go.sum
└── README.md
```

### Порты

- **40101** - Exchange 1
- **40102** - Exchange 2  
- **40103** - Exchange 3
- **8080** - MarketFlow API (если включен)

### Торговые пары

Система отслеживает следующие пары:
- BTCUSDT
- ETHUSDT
- SOLUSDT
- DOGEUSDT
- TONUSDT

## 🏗️ Архитектура

### Паттерны конкурентности

- **Fan-Out**: Каждый exchange слушается в отдельной горутине
- **Fan-In**: Все данные агрегируются в один канал
- **Worker Pool**: Обработка данных через пул воркеров
- **Generator**: Генерация тестовых данных

### Компоненты

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Exchange 1  │    │ Exchange 2  │    │ Exchange 3  │
│ :40101      │    │ :40102      │    │ :40103      │
└──────┬──────┘    └──────┬──────┘    └──────┬──────┘
       │                  │                  │
       └──────────────────┼──────────────────┘
                          │ (Fan-In)
                   ┌──────▼──────┐
                   │  MarketFlow │
                   │  Processor  │
                   └─────────────┘
                          │
                   ┌──────▼──────┐
                   │   Console   │
                   │   Output    │
                   └─────────────┘
```

## 🛠️ Отладка

### Проверка подключений

```bash
# Тестировать подключение к exchanges
nc -z localhost 40101
nc -z localhost 40102  
nc -z localhost 40103

# Или использовать make команду
make test-connection
```

### Логи

```bash
# Логи конкретного контейнера
docker-compose logs marketflow
docker-compose logs exchange1

# Логи всех контейнеров
docker-compose logs

# Следить за логами в реальном времени
docker-compose logs -f marketflow
```

### Проблемы и решения

**Проблема**: Контейнеры не запускаются
```bash
# Проверить статус
docker-compose ps

# Проверить логи
docker-compose logs
```

**Проблема**: Нет подключения к exchanges
```bash
# Проверить что exchanges запущены
docker-compose p

## 🎯 Особенности Live Mode

- ✅ Подключение к 3 exchanges одновременно
- ✅ Автоматическое переподключение при сбоях
- ✅ Парсинг различных форматов данных (JSON, текст)
- ✅ Вывод данных в реальном времени в консоль
- ✅ Graceful shutdown по SIGINT/SIGTERM
- ✅ Логирование всех событий
- ✅ Отказоустойчивость и failover

## 🔄 Режимы работы

**Live Mode** (текущая реализация):
- Подключение к реальным exchanges
- Обработка реальных рыночных данных
- Автоматическое переподключение

**Test Mode** (для будущей реализации):
- Генерация синтетических данных
- Тестирование без внешних зависимостей

---

## 👨🏻‍💻 Автор

- [![Status](https://img.shields.io/badge/alem-tishmal-success?logo=github)](https://platform.alem.school/git/tishmal) <a href="https://t.me/tim_shm" target="_blank"><img src="https://img.shields.io/badge/telegram-@tishmal-blue?logo=Telegram" alt="Status" /></a>