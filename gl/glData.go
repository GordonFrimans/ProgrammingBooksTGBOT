package gl

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Экспортируемые переменные (доступны из других пакетов через gl.AppID)
var (
	// main.go
	AppID       string
	AppHash     string
	PostgreURL  string
	SessionPath string

	// bot.go
	BotToken string

	// core.go
	DefaultSaveImage string

	// handelrs.go
	AdminID string
	LogPath string

	// fileOperation.go
	DefaultSaveBook string
)

func init() {
	// 1. Пытаемся найти .env
	// Сначала ищем в текущей папке, если нет — пробуем на уровень выше
	if err := loadEnv(".env"); err != nil {
		// Если не нашли в текущей, пробуем в родительской (для удобства запуска из вложенных папок)
		if err2 := loadEnv("../.env"); err2 != nil {
			fmt.Printf("⚠️ Внимание: файл .env не найден (%v). Надеемся, что переменные заданы в системе.\n", err)
			// Не делаем os.Exit(11), чтобы код мог работать в Docker, где файла .env может не быть физически
		}
	}

	// 2. ВАЖНО! Присваиваем значения переменным пакета ПОСЛЕ загрузки файла
	// Если этого не сделать, переменные gl.AppID останутся пустыми ""
	AppID = os.Getenv("appID")
	AppHash = os.Getenv("appHash")
	PostgreURL = os.Getenv("postgreURL")
	SessionPath = os.Getenv("sessionPath")

	BotToken = os.Getenv("botToken")
	DefaultSaveImage = os.Getenv("defaultSaveImage")
	AdminID = os.Getenv("adminID")
	LogPath = os.Getenv("logPath")
	DefaultSaveBook = os.Getenv("defaultSaveBook")

	// Можно добавить проверку критических переменных здесь
	if AppID == "" {
		fmt.Println("❌ Ошибка: AppID не задан!")
		os.Exit(1)
	}
}

func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")

		os.Setenv(key, value)
	}
	return scanner.Err()
}
