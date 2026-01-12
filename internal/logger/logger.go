package logger

import (
	"io"
	"log"
	"os"

	"HIGH_PR/gl"
)

var (
	// Logger - наш глобальный экземпляр логгера
	Logger  *log.Logger
	logFile *os.File // Переменная для хранения открытого файла
)

// SetupLogger инициализирует глобальный логгер
func SetupLogger(isDevMode bool) {
	var logOutput io.Writer
	var err error

	if isDevMode {
		// В режиме разработки пишем в консоль
		logOutput = os.Stdout
	} else {
		// В обычном режиме пишем в файл

		logFile, err = os.OpenFile(gl.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			log.Fatalf("Ошибка: не удалось открыть лог-файл: %v", err)
		}
		logOutput = logFile
	}

	// Создаем новый логгер
	Logger = log.New(logOutput, "INFO: ", log.LstdFlags|log.Lshortfile)
}

// Close закрывает файл лога, если он был открыт.
// Удобно вызывать через defer в main.
func Close() {
	if logFile != nil {
		err := logFile.Close()
		if err != nil {
			// Можно использовать стандартный логгер для этой критической ошибки
			log.Printf("Ошибка при закрытии лог-файла: %v\n", err)
		}
	}
}
