package bot

import (
	"HIGH_PR/gl"
	"HIGH_PR/internal/services"
	"context"
	"fmt"
	"log"

	//"strings"
	"strings"
	"sync"


	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// Bot — основная структура нашего приложения.
// Она содержит все необходимые зависимости для работы.
type Bot struct {
	// Внешние зависимости (передаются через New)
	client *telegram.Client // MTProto клиент для взаимодействия с Telegram API
	//bookService *services.BookService // Сервис для работы с книгами
	logger *log.Logger // Наш кастомный логгер

	// Внутреннее состояние
	dispatcher tg.UpdateDispatcher

	//Пул для работы с бд
	bookService *services.BookService
	mu   sync.Mutex
}

// WARNING
func New(
	client *telegram.Client,
	logger *log.Logger,
	dispatcher tg.UpdateDispatcher,
	bookService *services.BookService,

) *Bot {
	// Создаем новый диспетчер обновлений

	return &Bot{
		client:     client,
		logger:     logger,
		dispatcher: dispatcher,
		bookService:       bookService,
	}
}

func (b *Bot) Start(ctx context.Context) error {
	b.logger.Println("Регистрация обработчиков...")
	b.registerHandlers()

	return b.client.Run(ctx, func(ctx context.Context) error {
		// ⭐ ПРАВИЛЬНЫЙ ВАРИАНТ: Проверяем статус авторизации
		status, err := b.client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("проверка статуса: %w", err)
		}

		// Авторизуемся только если НЕ авторизованы
		if !status.Authorized {
			b.logger.Println("→ Выполняем авторизацию бота...")
			if _, err := b.client.Auth().Bot(ctx, gl.BotToken); err != nil {
				return fmt.Errorf("ошибка авторизации: %w", err)
			}
			b.logger.Println("✓ Авторизация успешна, сессия сохранена")
		} else {
			b.logger.Println("✓ Используем существующую сессию")
		}

		// Держим соединение открытым до отмены контекста
		<-ctx.Done()
		return ctx.Err()
	})
}



func (b *Bot) registerHandlers() {
	// Регистрируем обработчик для всех новых сообщений
	b.dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {

		// Передаем управление в наш основной маршрутизатор сообщений
		b.handleMessage(ctx, e, update)
		return nil

	})

	// Здесь можно регистрировать и другие обработчики:
	// b.dispatcher.OnBotCallbackQuery(...)
	// b.dispatcher.OnInlineQuery(...)
}

// handleMessage — основной маршрутизатор для текстовых сообщений.
func (b *Bot) handleMessage(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) {
	// Проверяем, что сообщение является текстовым
	msg, ok := update.Message.(*tg.Message)
	if !ok || msg.Out {
		return // Игнорируем не-сообщения или исходящие
	}

	b.logger.Printf("Получено сообщение: %s", msg.Message)

	text := msg.Message
	// Простая маршрутизация по тексту сообщения (команде)
	switch {
		case text == "/start":
			b.handleStart(ctx, e, msg)

		case strings.HasPrefix(text, "/add"):
			b.handleAddBook(ctx, e, msg,update)


		case text == "/show":
			b.handleShow(ctx, e, msg)

		case strings.HasPrefix(text, "/show_"):
			b.handleShowWithID(ctx,e,msg)

		case strings.HasPrefix(text, "/WithName"):
			b.handleShowWithName(ctx, e, msg)

		case text == "/help":
			// b.handleHelp(ctx, msg)

		case text == "/admin":
			b.handleAdmin(ctx, e, msg)

		case strings.HasPrefix(text, "/download_"):
			b.handleDownloadBook(ctx,e,msg)
		default:
			// b.handleSearch(ctx, msg)
	}
}
