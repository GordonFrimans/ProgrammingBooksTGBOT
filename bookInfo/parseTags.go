package bookinfo

import (
	"errors"
	"strings"
	"unicode"
)

// ParseMetadataFromTitle парсит заголовок и пытается извлечь Язык и Категорию.
// Возвращает: (language, tag, error)
func ParseMetadataFromTitle(title string) (string, string, error) {
	titleLower := strings.ToLower(title)

	var foundLang string
	var foundTag string

	// 1. Ищем Язык Программирования (проходим по карте CommonLanguages из tags.go)
	for alias, langName := range CommonLanguages {
		if containsWholeWord(titleLower, alias) {
			foundLang = langName
			break // Нашли язык - выходим. (Можно усложнить, если в названии 2 языка)
		}
	}

	// 2. Ищем Тэг/Тематику (проходим по карте TopicTags из tags.go)
	for alias, tagName := range TopicTags {
		if containsWholeWord(titleLower, alias) {
			foundTag = tagName
			break
		}
	}

	// 3. Логика "Умного заполнения"
	// Если нашли язык, но не нашли тэг — пытаемся подставить дефолтный для этого языка.
	// Например: Нашли "Go", но нет тэгов типа "Microservices". Ставим "Backend".
	if foundLang != "" && foundTag == "" {
		if defaultTag, ok := LanguageDefaultCategories[foundLang]; ok {
			foundTag = defaultTag
		}
	}

	// Если вообще ничего не нашли
	if foundLang == "" && foundTag == "" {
		return "", "", errors.New("тэги не найдены")
	}

	return foundLang, foundTag, nil
}

// containsWholeWord ищет word в text, учитывая границы слов.
// (Эта функция осталась без изменений, она отличная)
func containsWholeWord(text, word string) bool {
	startPos := 0
	for {
		// Ищем вхождение подстроки
		idx := strings.Index(text[startPos:], word)
		if idx == -1 {
			return false
		}

		// Вычисляем реальный индекс и конец слова
		realIdx := startPos + idx
		endIdx := realIdx + len(word)

		// Проверяем символ СЛЕВА
		isLeftBoundary := realIdx == 0 || !isWordChar(rune(text[realIdx-1]))

		// Проверяем символ СПРАВА
		isRightBoundary := endIdx == len(text) || !isWordChar(rune(text[endIdx]))

		// Если с обеих сторон границы — это целое слово!
		if isLeftBoundary && isRightBoundary {
			return true
		}

		// Если нашли часть слова (например "c" внутри "Apache"), ищем дальше
		startPos = realIdx + 1
	}
}

// isWordChar возвращает true, если символ — буква или цифра
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
