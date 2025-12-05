package bot

import (
	"HIGH_PR/gl"
	"context"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"strconv"
	"strings"
	"fmt"
	"github.com/gotd/td/telegram/uploader"
	"path/filepath"

)

func (b *Bot) handleCallback(

	ctx context.Context,
	e tg.Entities,
	update *tg.UpdateBotCallbackQuery,

) error {

	data := string(update.Data)
	switch  {
		case strings.HasPrefix(data, "page:"):
			pageStr := strings.TrimPrefix(data, "page:")
			targetPage, _ := strconv.Atoi(pageStr)

			// 1. Получаем книги
			books, err := b.bookService.GetAllBooks(ctx)
			if err != nil {
				_ = b.answerCallback(ctx, update.QueryID, "Ошибка получения данных")
				return err
			}

			// 2. Собираем текст и клаву
			newText, newKeyboard := b.buildBookPage(books, targetPage)

			// 3. Собираем InputPeer "вручную"
			var inputPeer tg.InputPeerClass

			switch p := update.Peer.(type) {
				case *tg.PeerUser:
					user, ok := e.Users[p.UserID]
					if !ok {
						return nil
					}
					inputPeer = &tg.InputPeerUser{
						UserID:     user.ID,
						AccessHash: user.AccessHash,
					}
				case *tg.PeerChat:
					inputPeer = &tg.InputPeerChat{
						ChatID: p.ChatID,
					}
				case *tg.PeerChannel:
					ch, ok := e.Channels[p.ChannelID]
					if !ok {
						return nil
					}
					inputPeer = &tg.InputPeerChannel{
						ChannelID:  ch.ID,
						AccessHash: ch.AccessHash,
					}
			}

			if inputPeer == nil {
				return nil
			}

			// 4. РЕДАКТИРУЕМ сообщение
			sender := message.NewSender(b.client.API())

			_, err = sender.
			To(inputPeer).          // -> *message.RequestBuilder (Builder)
			Markup(newKeyboard).    // Markup есть ТОЛЬКО тут [web:1]
			Edit(update.MsgID).     // -> *EditMessageBuilder
			StyledText(ctx, newText...) // редактируем текст и сущности [web:1][web:21]

			// 5. Убираем "часики"
			_ = b.answerCallback(ctx, update.QueryID, "")

			return err

		case data == "FileLog":

			sender := message.NewSender(b.client.API()).Peer(e, update)
			if err := b.handleButtonAllLog(ctx, sender); err != nil {
				return err
			}
		case data == "LastLog":

			sender := message.NewSender(b.client.API()).Peer(e, update)
			if err := b.handleLastLog(ctx, sender); err != nil {
				return err
			}

		case strings.HasPrefix(data, "download:"):
			sender := message.NewSender(b.client.API()).Peer(e, update)
			bookid := strings.TrimPrefix(data, "download:")
			id, _ := strconv.Atoi(bookid)
			filePath, err := b.bookService.GetFileBookWithID(ctx, id)
			if err != nil {
				b.logger.Println("Ошибка получения файла:", err)
				sender.Text(ctx, "Извините, файл не найден.")
				return err
			}

			// 4. Загружаем файл в Telegram
			// uploader.NewUploader разбивает файл на части и отправляет их
			u :=  uploader.NewUploader(b.client.API())

			b.logger.Println("Начинаю загрузку файла:", filePath)
			inputFile, err := u.FromPath(ctx, filePath)
			if err != nil {
				b.logger.Println("Ошибка загрузки (upload):", err)
				sender.Text(ctx, "Ошибка при загрузке файла.")
				return err
			}

			// 5. Подготавливаем InputMediaUploadedDocument [web:7][web:10]
			// Это ключевой момент: конвертируем загруженный файл в медиа-объект
			// Обязательно нужно указать имя файла через Attributes, иначе он придет как "file" без расширения
			fileName := filepath.Base(filePath)

			media := &tg.InputMediaUploadedDocument{
				File:     inputFile,
				MimeType: "application/pdf", // Желательно определять реально (например, "application/pdf")
				Attributes: []tg.DocumentAttributeClass{
					&tg.DocumentAttributeFilename{FileName: fileName}, // Чтобы у файла было имя
				},
				ForceFile: true, // Форсируем отправку именно как файл (документ)
			}

			// 6. Отправляем через метод Media(), а не Document()
			// Document() ждет tg.InputDocument (существующий файл), а мы шлем tg.InputMedia (новый контент)
			if _, err := sender.Media(ctx, message.Media(media)); err != nil {
				b.logger.Println("Ошибка отправки медиа:", err)
				sender.Text(ctx, fmt.Sprintf("Не удалось отправить файл: %s", err))
				return err
			}

			b.logger.Println("Файл успешно отправлен!")



		default:
			b.logger.Printf("Неизвестная кнопка: %s", string(data))
	}
	return b.answerCallback(ctx, update.QueryID, "Кнопка обработана!")




}




func (b *Bot) answerCallback(ctx context.Context, queryID int64, text string) error {
	_, err := b.client.API().MessagesSetBotCallbackAnswer(ctx, &tg.MessagesSetBotCallbackAnswerRequest{
		QueryID:   queryID,
		Message:   text,
		Alert:     false,
		CacheTime: 0,
	})
	return err
}

func (b *Bot) handleButtonAllLog(ctx context.Context, sender *message.RequestBuilder) error {
	b.SendFile(ctx, gl.LogPath, sender)
	return nil
}

func (b *Bot) handleLastLog(ctx context.Context, sender *message.RequestBuilder) error {
	// Твоя логика для кнопки "yes"
	b.SendLastLog(ctx, gl.LogPath, sender)
	return nil

}
