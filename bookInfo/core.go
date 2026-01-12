package bookinfo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"HIGH_PR/gl"
	"HIGH_PR/internal/logger"
)

type Result struct {
	Title       string
	Authors     []string
	Description string
	TextSnippet string
	Lang        string
	Img         string
}

type MinimalResponse struct {
	Items []struct {
		VolumeInfo struct {
			Title       string   `json:"title"`
			Authors     []string `json:"authors"`
			Description string   `json:"description"`
			Language    string   `json:"language"`
			ImageLinks  struct {
				Thumbnail      string `json:"thumbnail"`
				SmallThumbnail string `json:"smallThumbnail"`
			} `json:"imageLinks"`
		} `json:"volumeInfo"`
		SearchInfo struct {
			TextSnippet string `json:"textSnippet"`
		} `json:"searchInfo"`
	} `json:"items"`
}

func SearchBooks(name string) (Result, error) {
	name = strings.ReplaceAll(name, "_", " ")
	query := fmt.Sprintf("intitle:\"%s\" ru subject:\"computers\"", name)

	params := url.Values{}
	params.Add("q", query)
	params.Add("maxResults", "1")

	apiURL := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?%s", params.Encode())
	logger.Logger.Println("🤖🤖🤖 URL запроса к книге:", apiURL)

	// Retry логика: 3 попытки с задержкой 5 секунд
	const maxRetries = 3
	const retryDelay = 5 * time.Second

	var resp *http.Response
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err = http.Get(apiURL)

		// Если запрос прошёл успешно
		if err == nil {
			// Проверяем статус код
			if resp.StatusCode == http.StatusOK {
				// Успешный ответ - выходим из цикла
				break
			}

			// Если 503 - пробуем ретрай
			if resp.StatusCode == http.StatusServiceUnavailable {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				logger.Logger.Printf("Попытка %d/%d: Google API вернул 503: %s",
					attempt, maxRetries, string(body))

				// Если это не последняя попытка - делаем задержку
				if attempt < maxRetries {
					logger.Logger.Printf("Ожидание %v перед следующей попыткой...", retryDelay)
					time.Sleep(retryDelay)
					continue
				}

				// Последняя попытка провалилась
				return Result{}, fmt.Errorf("Google API недоступен после %d попыток: статус 503", maxRetries)
			}

			// Другие ошибки HTTP (не 503) - возвращаем сразу
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			logger.Logger.Printf("Google API вернул статус %d: %s", resp.StatusCode, string(body))
			return Result{}, fmt.Errorf("Google API вернул ошибку: статус %d", resp.StatusCode)
		}

		// Ошибка сетевого подключения
		logger.Logger.Printf("Попытка %d/%d: ошибка HTTP запроса: %v", attempt, maxRetries, err)

		if attempt < maxRetries {
			logger.Logger.Printf("Ожидание %v перед следующей попыткой...", retryDelay)
			time.Sleep(retryDelay)
			continue
		}

		return Result{}, fmt.Errorf("ошибка HTTP запроса после %d попыток: %w", maxRetries, err)
	}

	// Проверяем, что у нас валидный ответ
	if resp == nil || resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("не удалось получить данные после %d попыток", maxRetries)
	}

	defer resp.Body.Close()

	var minResp MinimalResponse
	if err := json.NewDecoder(resp.Body).Decode(&minResp); err != nil {
		return Result{}, fmt.Errorf("ошибка декодирования JSON: %w", err)
	}

	if len(minResp.Items) == 0 {
		return Result{}, fmt.Errorf("книга не найдена")
	}

	item := minResp.Items[0]

	result := Result{
		Title:       item.VolumeInfo.Title,
		Authors:     item.VolumeInfo.Authors,
		Description: item.VolumeInfo.Description,
		TextSnippet: item.SearchInfo.TextSnippet,
		Lang:        item.VolumeInfo.Language,
	}

	// Загружаем изображение
	imgURL := item.VolumeInfo.ImageLinks.Thumbnail
	if imgURL == "" {
		imgURL = item.VolumeInfo.ImageLinks.SmallThumbnail
	}
	result.Img = imgURL

	return result, nil
}

// DefaultSaveBook - константа для сохранения изображений + /img
func DownloadImage(imgURL string) (string, error) {
	imgURL = strings.Replace(imgURL, "zoom=1", "zoom=0", 1)
	// Шаг 1: Скачиваем изображение
	resp, err := http.Get(imgURL)
	if err != nil {
		return "", fmt.Errorf("ошибка при загрузке: %w", err)
	}
	defer resp.Body.Close()

	// Шаг 2: Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("статус %d при загрузке изображения", resp.StatusCode)
	}

	// Шаг 4: Генерируем уникальное имя файла (используем timestamp)
	filename := fmt.Sprintf("image_%d.jpg", time.Now().UnixNano())
	filePath := filepath.Join(gl.DefaultSaveImage, filename)

	// Шаг 5: Создаём файл для записи
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer file.Close()

	// Шаг 6: Копируем данные из ответа в файл
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка записи файла: %w", err)
	}

	// Возвращаем полный путь к сохранённому файлу
	return filePath, nil
}
