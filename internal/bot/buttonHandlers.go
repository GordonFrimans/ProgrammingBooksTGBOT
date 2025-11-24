package bot

import (
	"HIGH_PR/gl"
	"context"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

func (b *Bot) handleCallback(ctx context.Context, e tg.Entities, update *tg.UpdateBotCallbackQuery) error {
	data := update.Data
	userID := update.UserID

	b.logger.Printf("Кнопка нажата! Data: %s, UserID: %d", string(data), userID)

	// Создаем sender для ответа в тот же чат
	sender := message.NewSender(b.client.API()).Peer(e, update)

	switch string(data) {
	case "FileLog":
		// Обработка первой кнопки
		if err := b.handleButtonAllLog(ctx, sender); err != nil {
			return err
		}
	case "LastLog":
		// Обработка кнопки "Да"
		if err := b.handleLastLog(ctx, sender); err != nil {
			return err
		}
	default:
		b.logger.Printf("Неизвестная кнопка: %s", string(data))
	}

	// Отвечаем на callback query (убираем "часики" на кнопке)
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
