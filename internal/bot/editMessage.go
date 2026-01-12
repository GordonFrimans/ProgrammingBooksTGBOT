// Этот файл предназначен для красивого офрмеления сообщений! (если это не простое короткое сообщение тогда оно формируется в данном файле!)
package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	booktags "HIGH_PR/internal/repository/postgres/bookTags"

	"github.com/gotd/td/telegram/message/markup"

	"github.com/dustin/go-humanize"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

// buildBookPage собирает текст и кнопки для конкретной страницы
func (b *Bot) buildBookPage(books []booktags.BookWithTags, page int) ([]styling.StyledTextOption, tg.ReplyMarkupClass) {
	const booksPerPage = 3
	totalBooks := len(books)

	// 1. Считаем границы
	totalPages := (totalBooks + booksPerPage - 1) / booksPerPage
	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start := page * booksPerPage
	end := start + booksPerPage
	if end > totalBooks {
		end = totalBooks
	}

	// 2. Строим текст (тут твой код оформления)
	var text []styling.StyledTextOption
	text = append(text, styling.Plain(fmt.Sprintf("📖 Книги (Стр. %d/%d)\n\n", page+1, totalPages)))

	for i := start; i < end; i++ {
		var authors string
		if len(books[i].B.Authors) > 2 {
			authors = books[i].B.Authors[0] + ", " + books[i].B.Authors[1] + " ..."
		} else {
			authors = strings.Join(books[i].B.Authors, ", ")
		}

		text = append(text,
			styling.Bold("📚 Название: "),
			styling.Plain(books[i].B.Title+"\n\n"),

			styling.Bold("👨‍💼 Авторы: "),
			styling.Plain(authors+"\n\n"),

			styling.Bold("📝 Описание:\n"),
			styling.Italic("    "+books[i].B.TextSnippet+"\n\n"),

			styling.Custom(func(eb *entity.Builder) error {
				eb.Format("🔗 Скачать:", entity.Bold())
				return nil
			}),
			styling.Plain(fmt.Sprintf(" /download_%d\n", books[i].B.ID)),

			styling.Custom(func(eb *entity.Builder) error {
				eb.Format("🔎 Подробнее:", entity.Bold())
				return nil
			}),
			styling.Plain(fmt.Sprintf(" /show_%d\n", books[i].B.ID)),

			styling.Plain("\n"),
			styling.Plain("━━━━━━━━━━"),
			styling.Plain("\n\n"),
		)
	}

	// 3. Строим кнопки. ВНИМАНИЕ: мы сразу пишем номер НУЖНОЙ страницы
	var rows []tg.KeyboardButtonClass

	// Если не первая страница -> кнопка "Назад" ведет на (page - 1)
	if page > 0 {
		rows = append(rows, markup.Callback("⬅️ Назад", []byte(fmt.Sprintf("page:%d", page-1))))
	}

	// Если не последняя страница -> кнопка "Вперед" ведет на (page + 1)
	if page < totalPages-1 {
		rows = append(rows, markup.Callback("Вперед ➡️", []byte(fmt.Sprintf("page:%d", page+1))))
	}
	if len(rows) == 0 {
		return text, nil
	}

	return text, markup.InlineRow(rows...)
}

// БЫЛО: func ShowBooksMessage(ctx context.Context, msg *message.RequestBuilder, pool *pgxpool.Pool)
// СТАЛО:
func (b *Bot) ShowBooksMessage(ctx context.Context, msg *message.RequestBuilder, books []booktags.BookWithTags) error {
	// Показываем самую первую страницу (0)
	text, keyboard := b.buildBookPage(books, 0)

	_, err := msg.Markup(keyboard).StyledText(ctx, text...)
	return err
}

func calculatePageCount(totalBooks, booksPerPage int) int {
	return (totalBooks + booksPerPage - 1) / booksPerPage
}

func formatAuthors(authors []string) string {
	return strings.Join(authors, ", ")
}

func ShowBookWithIDMessage(ctx context.Context, client *telegram.Client, peer tg.InputPeerClass, book booktags.BookWithTags) error {
	// 1. Инициализируем Uploader
	uploader := uploader.NewUploader(client.API())
	inpF, err := uploader.FromPath(ctx, book.B.Img)
	if err != nil {
		return fmt.Errorf("upload error: %w", err)
	}

	// 2. Подготовка переменных
	randomID := time.Now().UnixNano()
	media := &tg.InputMediaUploadedPhoto{File: inpF}
	fileSize := humanize.Bytes(uint64(book.B.FileSize))
	addedAt := book.B.AddedAt.Format("02.01.2006 15:04")

	// Обрезаем описание заранее
	desc := book.B.Description
	// Преобразуем строку в срез рун (символов)
	descRunes := []rune(desc)

	// Проверяем длину именно в символах
	if len(descRunes) > 600 {
		// Безопасно обрезаем по символам и собираем обратно в строку
		desc = string(descRunes[:597]) + "..."
	}

	// 3. ИСПОЛЬЗУЕМ entity.Builder ВМЕСТО styling
	// Builder сам посчитает все смещения (offsets/lengths)
	// ЯП
	var langProg string
	if book.T.ProgrammingLang[0] == "" {
		langProg = "-"
	} else {
		langProg = book.T.ProgrammingLang[0]
	}

	// Язык книги
	langMap := map[string]string{
		"ru": "🇷🇺 Русский",
		"en": "🇬🇧 Английский",
		// добавь другие по необходимости
	}

	var b entity.Builder

	b.Bold("📚 Название: ")
	b.Plain(book.B.Title + "\n\n")

	b.Bold("👨‍💼 Авторы: ")
	b.Plain(formatAuthors(book.B.Authors) + "\n\n")

	b.Bold("📝 Описание:\n")
	b.Italic("    " + desc + "\n\n")

	b.Bold("💾 Размер файла: ")
	b.Plain(fileSize + "\n")

	b.Bold("🗂 Тип файла: ")
	b.Plain(book.B.FileType + "\n")

	b.Bold("📅 Добавлено: ")
	b.Plain(addedAt + "\n")

	b.Bold("⬇️ Скачиваний: ")
	b.Plain(fmt.Sprintf("%d\n", book.B.DownloadCount))

	b.Bold("🌐 Язык: ")
	b.Plain(langMap[book.T.Lang] + "\n")

	b.Bold("💻 ЯП: ")
	b.Plain(langProg + "\n")

	b.Bold("🏷️ Категория: ")
	b.Plain(book.T.OtherTag[0] + "\n")

	// 4. Генерируем готовый текст и entities
	captionText, entities := b.Complete()

	keyboard := markup.InlineRow(markup.Callback("Скачать", []byte(fmt.Sprintf("download:%d", book.B.ID))))

	// 5. Отправляем
	_, err = client.API().MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:        peer,
		Message:     captionText,
		RandomID:    randomID,
		Media:       media,
		Entities:    entities, // Теперь тип совпадает ([]tg.MessageEntityClass)
		ReplyMarkup: keyboard,
	})

	return err
}

// SendHelpMessage отправляет красиво оформленное сообщение с помощью
func (b *Bot) SendHelpMessage(ctx context.Context) []styling.StyledTextOption {
	// Мы используем styling.Plain, styling.Bold, styling.Code и styling.Italic
	// чтобы собрать сообщение как конструктор.
	var text []styling.StyledTextOption
	text = append(text,
		// Заголовок
		styling.Bold("🤖 Добро пожаловать! Вот что я умею:\n\n"),

		// --- Основные команды ---
		styling.Plain("🚀 "),
		styling.Plain("/start"),
		styling.Plain(" — Начать работу с ботом.\n\n"),

		styling.Plain("📚 "),
		styling.Plain("/show"),
		styling.Plain(" — Показать список всех доступных книг.\n\n"),

		// --- Работа с конкретной книгой ---
		styling.Plain("🔍 "),
		styling.Code("/show_num"),
		styling.Plain(" — Показать подробную информацию о книге по её ID.\n"),
		styling.Italic("Пример: /show_1\n\n"),

		styling.Plain("⬇️ "),
		styling.Code("/download_num"),
		styling.Plain(" — Скачать файл книги по её ID.\n"),
		styling.Italic("Пример: /download_2\n\n"),

		// --- Добавление книг ---
		styling.Bold("📥 Добавление книг:\n"),
		styling.Plain("Просто отправь мне файл с командой в подписи:\n\n"),

		styling.Plain("1️⃣ "),
		styling.Code("/add"),
		styling.Plain(" — Автоматическое добавление.\n"),
		styling.Italic("(прикрепите файл книги к этому сообщению)\n\n"),

		styling.Plain("2️⃣ "),
		styling.Code("/add <Название>"),
		styling.Plain(" — Добавление с указанием названия вручную.\n"),
		styling.Italic("Используй, если автоматика ошиблась.\n"),
		styling.Italic("(прикрепите файл книги к этому сообщению)"),
	)
	return text
}
