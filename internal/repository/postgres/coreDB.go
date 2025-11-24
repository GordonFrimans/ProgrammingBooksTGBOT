package postgres

import (
	"context"
	"fmt"
	"time"

	"HIGH_PR/internal/logger"
	booktags "HIGH_PR/internal/repository/postgres/bookTags"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDB создаёт и настраивает connection pool
func InitDB(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Парсим конфиг
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфига БД: %w", err)
	}

	// Настраиваем параметры pool
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 15 * time.Minute

	// Создаём pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	logger.Logger.Println("✓ Подключение к БД успешно установлено")
	return pool, nil
}



// Setup инициализирует БД: создаёт pool и таблицы
func Setup(databaseURL string) (*pgxpool.Pool, error) {
	// 1. Создаём pool
	pool, err := InitDB(databaseURL)
	if err != nil {
		return nil, err
	}

	// 2. Создаём таблицы
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := booktags.CreateTables(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
