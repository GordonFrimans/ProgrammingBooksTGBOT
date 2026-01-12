package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"HIGH_PR/gl"
	"HIGH_PR/internal/bot"
	"HIGH_PR/internal/logger"
	"HIGH_PR/internal/middleware"
	"HIGH_PR/internal/repository/postgres"
	booktags "HIGH_PR/internal/repository/postgres/bookTags"
	"HIGH_PR/internal/services"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func init() {
	createFolders()
	readEnv()
}

func createFolders() {
	dirsToCreate := []string{
		gl.DefaultSaveBook,
		gl.DefaultSaveImage,
	}

	// Добавляем папку логов (если путь задан)
	if gl.LogPath != "" {
		// filepath.Dir корректно отрежет "app.log" и оставит "log"
		// Не забудь import "path/filepath"
		dirsToCreate = append(dirsToCreate, filepath.Dir(gl.LogPath))
	}

	// Добавляем папку сессии
	if gl.SessionPath != "" {
		dirsToCreate = append(dirsToCreate, filepath.Dir(gl.SessionPath))
	}

	for _, dir := range dirsToCreate {
		// filepath.Clean убирает лишние слэши и точки
		cleanDir := filepath.Clean(dir)

		// Skip, если путь пустой или "." (текущая папка)
		if cleanDir == "." || cleanDir == "" {
			continue
		}

		// 0755 - стандартные права (rwxr-xr-x)
		err := os.MkdirAll(cleanDir, 0o755)
		if err != nil {
			fmt.Printf("❌ Критическая ошибка: Не могу создать папку '%s': %v\n", cleanDir, err)
			// Тут лучше упасть, так как без папок бот работать не сможет
			os.Exit(1)
		}
		fmt.Printf("✅ Папка проверена/создана: %s\n", cleanDir)
	}
}

func readEnv() {
	// 1. Определяем флаг -dev
	devMode := flag.Bool("dev", false, "Активировать режим разработки (лог в консоли)")
	flag.Parse()

	// 2. СРАЗУ настраиваем наш кастомный логгер
	logger.SetupLogger(*devMode)

	// 3. А ТЕПЕРЬ используем ТОЛЬКО его для всех сообщений
	if *devMode {
		logger.Logger.Println("Режим разработки активирован.")
	} else {
		// Это сообщение теперь пойдет в файл, как и ожидалось
		logger.Logger.Println("Обычный режим работы.")
	}
}

func main() {
	logger.Logger.Println("🚦 Запуск")
	defer logger.Close()
	// 2. Создаём диспетчер ЗАРАНЕЕ
	dispatcher := tg.NewUpdateDispatcher()
	// ATTENTION
	appID, _ := strconv.Atoi(gl.AppID)

	// 3. Создаём клиент, передавая ему диспетчер через Options
	client := telegram.NewClient(appID, gl.AppHash, telegram.Options{
		UpdateHandler: dispatcher,
		// ⭐ ДОБАВЛЯЕМ MIDDLEWARE ЗДЕСЬ
		Middlewares: []telegram.Middleware{
			middleware.LoggingMiddleware(logger.Logger),
			// Здесь можно добавить больше middleware
		},
		SessionStorage: &session.FileStorage{
			Path: gl.SessionPath,
		},
	})
	pool, err := postgres.Setup(gl.PostgreURL)
	if err != nil {
		logger.Logger.Println("Ошибка подкл к DB: ", err)
		os.Exit(5432)
	}

	bookRep := booktags.NewBookRepository(pool)
	bookService := services.NewBookService(bookRep)

	botApp := bot.New(client, logger.Logger, dispatcher, bookService)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := botApp.Start(ctx); err != nil {
		logger.Logger.Fatalf("Ошибка при запуске: %v", err)
	}

	logger.Logger.Println("Бот остановлен.")
}
