package booktags

import "time"

// Book представляет книгу из базы данных
type Book struct {
	ID int

	Title string

	Authors []string

	Description string

	TextSnippet string

	Img string

	FilePath string
	FileSize int64 // Кол-во байт
	FileType string

	AddedBy string
	AddedAt time.Time

	DownloadCount int // Кол-во скачиваний
}

// Tag представляет тег из базы данных
type Tag struct {
	ID int

	BookID int

	OtherTag        []string
	Lang            string
	ProgrammingLang []string
}

// BookWithTags — удобная структура для получения книги со всеми тегами
type BookWithTags struct {
	B Book
	T Tag
}
