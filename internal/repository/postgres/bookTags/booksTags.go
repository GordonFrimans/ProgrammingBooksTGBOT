package booktags

import (
	"HIGH_PR/internal/logger"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTables создаёт таблицы, если они не существуют
func CreateTables(ctx context.Context, pool *pgxpool.Pool) error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS books (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		authors VARCHAR(100)[],
		description TEXT,
		textsnippet TEXT,
		img VARCHAR(500),
		file_path VARCHAR(500),
		file_size BIGINT,
		file_type VARCHAR(10),
		added_by VARCHAR(50),
		added_at TIMESTAMP DEFAULT NOW(),
		download_count INT DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);

		CREATE TABLE IF NOT EXISTS tags (
			id SERIAL PRIMARY KEY,
			book_id INT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
			lang VARCHAR(20),
			programming_lang VARCHAR(100)[],
			other_tag VARCHAR(100)[],
			UNIQUE(book_id)
			);

			CREATE INDEX IF NOT EXISTS idx_tags_book ON tags(book_id);
			CREATE INDEX IF NOT EXISTS idx_tags_other ON tags USING GIN (other_tag);
			CREATE INDEX IF NOT EXISTS idx_tags_programming_lang ON tags USING GIN (programming_lang);
			CREATE INDEX IF NOT EXISTS idx_tags_lang ON tags(lang);
			`

	_, err := pool.Exec(ctx, createTableQuery)
	if err != nil {
		logger.Logger.Println("Не удалось создать таблицы:", err)
		return err
	}

	logger.Logger.Println("Таблицы успешно созданы/проверены")
	return nil
}
