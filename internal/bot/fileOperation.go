package bot

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"HIGH_PR/gl"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

// Универскальная костыльная функция для отправки файла (без возможности добавить caption) (Но с возможностью указать аттрибуты файла такие как MIME и тп...)
// WARNING
func (b *Bot) SendFile(ctx context.Context, path string, sender *message.RequestBuilder) {
	upload := sender.Upload(message.FromPath(path))
	inputFile, err := upload.AsInputFile(ctx)
	if err != nil {
		b.logger.Printf("Ошибка создания (inputFile): %v", err)
	}
	_, err = sender.Media(ctx,
		message.UploadedDocument(inputFile).
			MIME("text/plain").
			Filename("BOT_LOG.log"))
	if err != nil {
		b.logger.Printf("Ошибка отправки: %v", err)
	}
}

func (b *Bot) SendLastLog(ctx context.Context, path string, sender *message.RequestBuilder) {
	lastStrLog, err := ReadLastLinesAsString(path, 20)
	if err != nil {
		b.logger.Println("Ошибка чтения файла с логами!")
	}
	_, err = sender.StyledText(ctx, styling.Code(lastStrLog))
	if err != nil {
		b.logger.Println("Ошибка отпраки: ", err)
	}
}

func ReadLastLinesAsString(path string, n int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// Читаем все строки
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	// Берём последние N строк
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	// Объединяем в одну строку с переносами строк
	return strings.Join(lines, "\n"), nil
}

func (b *Bot) DownloadFile(ctx context.Context, media *tg.MessageMediaDocument) error {
	b.logger.Println("Запуск загрузки файла!")
	doc, ok := media.Document.(*tg.Document)
	if !ok {
		return fmt.Errorf("не удалось получить документ")
	}

	// Шаг 2: Создать InputFileLocation для загрузки
	location := &tg.InputDocumentFileLocation{
		ID:            doc.ID,
		AccessHash:    doc.AccessHash,
		FileReference: doc.FileReference,
		ThumbSize:     "", // пустая строка = основной файл (не превью)
	}

	// Шаг 3: Определить путь сохранения
	filename := GetDocumentName(doc) // твоя функция
	savePath := filepath.Join(gl.DefaultSaveBook, filename)

	b.logger.Printf("📥 Загрузка: %s", filename)

	// Шаг 4: Загрузить файл
	_, err := downloader.NewDownloader().
		Download(b.client.API(), location).
		ToPath(ctx, savePath)
	if err != nil {
		return fmt.Errorf("ошибка загрузки: %w", err)
	}

	b.logger.Printf("✅ Файл сохранён: %s", savePath)
	return nil
}

// Вспомогательная функция для получения имени документа
func GetDocumentName(doc *tg.Document) string {
	for _, attr := range doc.Attributes {
		if fn, ok := attr.(*tg.DocumentAttributeFilename); ok {
			return fn.FileName
		}
	}
	return "document.pdf"
}

func DeleteType(name string) string {
	res := strings.Replace(name, ".pdf", "", 1)
	return res
}

// Простая и надёжная функция извлечения формата из имени файла
func ExtractFileFormat(filename string) string {
	// filepath.Ext возвращает расширение с точкой, например ".pdf"
	ext := filepath.Ext(filename)

	// Убираем точку и приводим к нижнему регистру
	format := strings.ToLower(strings.TrimPrefix(ext, "."))

	// Если расширения нет, возвращаем дефолтное значение
	if format == "" {
		return "unknown"
	}

	return format
}
