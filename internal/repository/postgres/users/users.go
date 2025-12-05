package users

import (
	"HIGH_PR/internal/logger"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Создание таблицы хранящий статистику за день!
func CreateTableUsers(ctx context.Context, pool *pgxpool.Pool) error {
	logger.Logger.Println("Создане таблицы статистики")
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS daily_stats (

		date DATE PRIMARY KEY -- Дата

		unique_users_count INT DEFAULT 0 --Кол-во уникальных пользователей посетивших бота за указ дату

		message_count INT DEFAULT 0 --Кол-во сообщений

		)


		-- Индекс для сортировки по дате (для графиков/отчетов)
		CREATE INDEX IF NOT EXISTS idx_daily_stats_date ON daily_stats(date DESC);
		`
	_, err := pool.Exec(ctx, createTableQuery)
	if err != nil {
		logger.Logger.Println("Не удалось создать таблицы:", err)
		return err
	}

	logger.Logger.Println("Таблицы успешно созданы/проверены")
	return nil

}
