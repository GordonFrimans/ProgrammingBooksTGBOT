package middleware

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// LoggingMiddleware логирует все RPC-вызовы к Telegram API
func LoggingMiddleware(logger *log.Logger) telegram.Middleware {
	return telegram.MiddlewareFunc(func(next tg.Invoker) telegram.InvokeFunc {
		return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
			start := time.Now()

			// Получаем название метода
			methodName := extractMethodName(input)

			logger.Printf("→ Вызов метода: %s", methodName)

			// Вызываем следующий обработчик в цепочке
			err := next.Invoke(ctx, input, output)

			duration := time.Since(start)

			if err != nil {
				logger.Printf("❌ Метод %s завершился с ошибкой за %v: %v",
					methodName, duration, err)
			} else {
				logger.Printf("✅ Метод %s выполнен успешно за %v",
					methodName, duration)
			}

			return err
		}
	})
}

// extractMethodName извлекает имя метода из запроса
func extractMethodName(input bin.Encoder) string {
	// Сначала пробуем TypeName (более читаемое имя)
	if typed, ok := input.(interface{ TypeName() string }); ok {
		return typed.TypeName()
	}

	// Если TypeName недоступен, используем TypeID
	if typed, ok := input.(interface{ TypeID() uint32 }); ok {
		return fmt.Sprintf("0x%x", typed.TypeID())
	}

	return "unknown"
}
