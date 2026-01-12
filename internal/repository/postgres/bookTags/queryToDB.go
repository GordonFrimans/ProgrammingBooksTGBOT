package booktags

import (
	"context"
	"errors"
	"fmt"

	bookinfo "HIGH_PR/bookInfo"
	"HIGH_PR/internal/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BookRepository инкапсулирует работу с книгами в БД
type BookRepository struct {
	pool *pgxpool.Pool
}

// NewBookRepository создаёт новый репозиторий
func NewBookRepository(pool *pgxpool.Pool) *BookRepository {
	return &BookRepository{pool: pool}
}

// GetAllBooks возвращает все книги с тегами
func (r *BookRepository) GetAllBooks(ctx context.Context) ([]BookWithTags, error) {
	logger.Logger.Println("Запрос для получения всех книг")

	sqlQuery := `SELECT
	b.id,
	b.title,
	b.authors,
	b.description,
	b.textSnippet,
	b.img,
	b.file_path,
	b.file_size,
	b.file_type,
	b.added_by,
	b.added_at,
	b.download_count,
	t.id,
	t.book_id,
	t.other_tag,
	t.lang,
	t.programming_lang
	FROM books b
	JOIN tags t ON b.id = t.book_id`

	rows, err := r.pool.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []BookWithTags
	for rows.Next() {
		bt := BookWithTags{}
		err = rows.Scan(
			&bt.B.ID, &bt.B.Title, &bt.B.Authors, &bt.B.Description, &bt.B.TextSnippet, &bt.B.Img,
			&bt.B.FilePath, &bt.B.FileSize, &bt.B.FileType,
			&bt.B.AddedBy, &bt.B.AddedAt, &bt.B.DownloadCount,
			&bt.T.ID, &bt.T.BookID, &bt.T.OtherTag,
			&bt.T.Lang, &bt.T.ProgrammingLang,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, bt)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

func (r *BookRepository) AddBook(ctx context.Context, bt BookWithTags) error {
	// Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Откатываем если не закоммитим
	// Проверка на наличие по (title)
	var existingTitel string
	checkQuery := `SELECT id FROM books WHERE title = $1 LIMIT 1`
	err = tx.QueryRow(ctx, checkQuery, bt.B.Title).Scan(&existingTitel)
	if err == nil {
		return fmt.Errorf("Книга с таким именем уже есть!")
	}
	// Загрузка изображения
	imgData, err := bookinfo.DownloadImage(bt.B.Img)
	if err != nil {
		logger.Logger.Printf("Не удалось загрузить обложку: %v", err)
		return err
	} else {
		bt.B.Img = imgData
	}
	//===================================

	// 1. Вставляем книгу
	query := `
	INSERT INTO books (title, authors, description, textSnippet, img, file_path, file_size, file_type, added_by, added_at, download_count)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,$10,$11)
	RETURNING id`

	var bookID int
	err = tx.QueryRow(ctx, query,
		bt.B.Title,
		bt.B.Authors,
		bt.B.Description,
		bt.B.TextSnippet,
		bt.B.Img,
		bt.B.FilePath,
		bt.B.FileSize,
		bt.B.FileType,
		bt.B.AddedBy,
		bt.B.AddedAt,
		bt.B.DownloadCount,
	).Scan(&bookID)
	if err != nil {
		return fmt.Errorf("failed to insert book: %w", err)
	}

	// 2. Вставляем теги
	tagQuery := `
	INSERT INTO tags (book_id, other_tag, lang, programming_lang)
	VALUES ($1, $2, $3, $4)`

	_, err = tx.Exec(ctx, tagQuery,
		bookID,
		bt.T.OtherTag,
		bt.T.Lang,
		bt.T.ProgrammingLang,
	)
	if err != nil {
		return fmt.Errorf("failed to insert tags: %w", err)
	}

	// Коммитим транзакцию
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *BookRepository) BookWithID(ctx context.Context, id int) (BookWithTags, error) {
	logger.Logger.Printf("Запрос на получение книги |ID=%d|\n", id)

	sqlQuery := `
	SELECT
	b.id, b.title, b.authors, b.description, b.textSnippet, b.img,
	b.file_path, b.file_size, b.file_type,
	b.added_by, b.added_at, b.download_count,
	t.id, t.book_id, t.other_tag,
	t.lang, t.programming_lang
	FROM books b
	JOIN tags t ON b.id = t.book_id
	WHERE b.id = $1
	LIMIT 1
	`

	var bt BookWithTags
	err := r.pool.QueryRow(ctx, sqlQuery, id).Scan(
		&bt.B.ID, &bt.B.Title, &bt.B.Authors, &bt.B.Description, &bt.B.TextSnippet, &bt.B.Img,
		&bt.B.FilePath, &bt.B.FileSize, &bt.B.FileType,
		&bt.B.AddedBy, &bt.B.AddedAt, &bt.B.DownloadCount,
		&bt.T.ID, &bt.T.BookID, &bt.T.OtherTag,
		&bt.T.Lang, &bt.T.ProgrammingLang,
	)
	if err != nil {
		return BookWithTags{}, err
	}

	return bt, nil
}

func (r *BookRepository) GetFileBookWithID(ctx context.Context, id int) (string, error) {
	// Логгер лучше использовать на уровне повыше (в usecase), но в учебных целях ок.
	// logger.Logger.Println("Получаем путь к файлу книги по ID")

	// Лучше писать запрос в одну строку или использовать константы,
	// но такой формат тоже читаем.
	const sqlQuery = `SELECT file_path FROM books WHERE id = $1`

	var filePath string
	err := r.pool.QueryRow(ctx, sqlQuery, id).Scan(&filePath)
	if err != nil {
		// Проверяем, что ошибка именно "ничего не найдено"
		if errors.Is(err, pgx.ErrNoRows) {
			// Можно вернуть спец. ошибку или пустую строку, зависит от логики
			return "", fmt.Errorf("book with id %d not found: %w", id, pgx.ErrNoRows)
		}
		// Оборачиваем остальные ошибки для контекста
		return "", fmt.Errorf("repository.GetFileBookWithID: %w", err)
	}

	return filePath, nil
}

func (r *BookRepository) AddDownloadCountWithID(ctx context.Context, id int) error {
	sqlQuery := `
	UPDATE books
	SET download_count = download_count + 1
	WHERE id = $1
	`

	// Выполняем запрос
	commandTag, err := r.pool.Exec(ctx, sqlQuery, id)
	if err != nil {
		return fmt.Errorf("не удалось обновить счётчик: %w", err)
	}

	// Проверяем, что запись действительно обновилась
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("книга с id=%d не найдена", id)
	}

	return nil
}

// ShowBooksWithTag ищет книги, где переданный тег встречается
// либо в programming_lang, либо в other_tag.
func (r *BookRepository) ShowBooksWithTag(ctx context.Context, tag string) ([]BookWithTags, error) {
	// Мы ищем книги, где массив programming_lang СОДЕРЖИТ tag
	// ИЛИ массив other_tag СОДЕРЖИТ tag.
	// Используем оператор && (пересечение) и приводим наш одиночный тег к массиву ARRAY[$1].
	sqlQuery := `
	SELECT
	b.id, b.title, b.authors, b.description, b.textSnippet, b.img,
	b.file_path, b.file_size, b.file_type,
	b.added_by, b.added_at, b.download_count,
	t.id, t.book_id, t.other_tag,
	t.lang, t.programming_lang
	FROM books b
	JOIN tags t ON b.id = t.book_id
	WHERE
	t.programming_lang && ARRAY[$1]::varchar[]
	OR
	t.other_tag && ARRAY[$1]::varchar[]
	`

	// Передаем tag один раз, он подставится в $1
	rows, err := r.pool.Query(ctx, sqlQuery, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []BookWithTags

	for rows.Next() {
		var bt BookWithTags
		// ВАЖНО: порядок переменных должен строго соответствовать SELECT
		err = rows.Scan(
			&bt.B.ID, &bt.B.Title, &bt.B.Authors, &bt.B.Description, &bt.B.TextSnippet, &bt.B.Img,
			&bt.B.FilePath, &bt.B.FileSize, &bt.B.FileType,
			&bt.B.AddedBy, &bt.B.AddedAt, &bt.B.DownloadCount,
			&bt.T.ID, &bt.T.BookID, &bt.T.OtherTag,
			&bt.T.Lang, &bt.T.ProgrammingLang,
		)
		if err != nil {
			return nil, err
		}
		books = append(books, bt)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

func (r *BookRepository) SearchBooksWithTitleDesc(ctx context.Context, query string) ([]BookWithTags, error) {
	// В функции инициализации репозитория или main.go
	_, err := r.pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS pg_trgm")
	if err != nil {
		// Логируем ошибку, но обычно прав доступа может не хватить
		return nil, fmt.Errorf("failed to set similarity threshold: %w", err)
	}

	// 1. Настройка порога чувствительности для текущей транзакции/сессии
	// 0.3-0.4 — хороший баланс. Чем выше число, тем строже поиск.
	_, err = r.pool.Exec(ctx, "SET pg_trgm.word_similarity_threshold = 0.3")
	if err != nil {
		return nil, fmt.Errorf("failed to set similarity threshold: %w", err)
	}

	// 2. SQL запрос с ранжированием
	sqlQuery := `
	SELECT
	b.id, b.title, b.authors, b.description, b.textSnippet, b.img,
	b.file_path, b.file_size, b.file_type,
	b.added_by, b.added_at, b.download_count,
	NULL::int, NULL::int, NULL::varchar[], NULL::varchar, NULL::varchar[]
	FROM books b
	WHERE
	($1::text <% b.title) OR ($1::text <% b.description)
	ORDER BY
	-- Здесь тоже лучше явно привести к text, хотя функции обычно умнее операторов
	GREATEST(word_similarity($1::text, b.title), word_similarity($1::text, b.description) * 0.6) DESC
	LIMIT 50;
	`

	rows, err := r.pool.Query(ctx, sqlQuery, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []BookWithTags

	for rows.Next() {
		var bt BookWithTags
		// Временные переменные для сканирования NULL-полей тегов (чтобы Scan не упал)
		var tID, tBookID *int
		var tOtherTag, tProgLang []string
		var tLang *string

		err = rows.Scan(
			&bt.B.ID, &bt.B.Title, &bt.B.Authors, &bt.B.Description, &bt.B.TextSnippet, &bt.B.Img,
			&bt.B.FilePath, &bt.B.FileSize, &bt.B.FileType,
			&bt.B.AddedBy, &bt.B.AddedAt, &bt.B.DownloadCount,
			&tID, &tBookID, &tOtherTag, &tLang, &tProgLang,
		)
		if err != nil {
			return nil, err
		}
		// Можно заполнить bt.T если нужно, но здесь они пустые
		books = append(books, bt)
	}

	return books, nil
}

// В будущем здесь будут другие методы:
// func (r *BookRepository) GetByID(ctx context.Context, id int) (*Book, error) { ... }
// func (r *BookRepository) Create(ctx context.Context, book *Book) (int, error) { ... }
// func (r *BookRepository) Delete(ctx context.Context, id int) error { ... }
