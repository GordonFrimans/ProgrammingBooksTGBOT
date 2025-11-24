// Этот файл предназначен для красивого офрмеления сообщений! (если это не простое короткое сообщение тогда оно формируется в данном файле!)
package bot

import (
	"HIGH_PR/internal/logger"
	"HIGH_PR/internal/repository/postgres/bookTags"
	"context"
	"fmt"

	"github.com/gotd/td/telegram/message"
	//"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/telegram/message/entity"
)

import (
	"github.com/gotd/td/telegram/message/styling"
	"strings"
	"github.com/gotd/td/telegram/uploader"
	"time"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/dustin/go-humanize"


)

// БЫЛО: func ShowBooksMessage(ctx context.Context, msg *message.RequestBuilder, pool *pgxpool.Pool)
// СТАЛО:
func ShowBooksMessage(ctx context.Context, msg *message.RequestBuilder, books []booktags.BookWithTags) error {
	logger.Logger.Println("Создаем сообщение с книгами!")



	var bookPages [][]styling.StyledTextOption
	totalBooks := len(books) // <-- используем переданный аргумент
	booksPerPage := 5
	countPage := calculatePageCount(totalBooks, booksPerPage)

	for page := 0; page < countPage; page++ {
		var styledTexts []styling.StyledTextOption
		start := page * booksPerPage
		end := start + booksPerPage

		if end > totalBooks {
			end = totalBooks
		}
		styledTexts = append(styledTexts, styling.Plain("━━━━━━━━━━\n\n"))
		for i := start; i < end; i++ {
			var authors string
			if len(books[i].B.Authors) > 2 {
				authors = books[i].B.Authors[0] + ", " + books[i].B.Authors[1] + " ..."
			} else {
				authors = strings.Join(books[i].B.Authors, ", ")
			}

			styledTexts = append(styledTexts,
					     styling.Bold("📚 Название: "),
					     styling.Plain(books[i].B.Title + "\n\n"),

					     styling.Bold("👨‍💼 Авторы: "),
					     styling.Plain(authors + "\n\n"),

					     styling.Bold("📝 Описание:\n"),
					     styling.Italic("    " + books[i].B.TextSnippet + "\n\n"),

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




		bookPages = append(bookPages, styledTexts)
	}

	_, err := msg.StyledText(ctx, bookPages[0]...)
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

	b.Bold("🏷️ Язык: ")
	b.Plain(book.T.Lang + "\n")

	// 4. Генерируем готовый текст и entities
	captionText, entities := b.Complete()



	// 5. Отправляем
	_, err = client.API().MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     peer,
		Message:  captionText,
		RandomID: randomID,
		Media:    media,
		Entities: entities, // Теперь тип совпадает ([]tg.MessageEntityClass)
	})

	return err
}




