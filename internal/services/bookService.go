package services

import (
	booktags "HIGH_PR/internal/repository/postgres/bookTags"
	"context"
)

// BookService содержит бизнес-логику для работы с книгами
type BookService struct {
	repo *booktags.BookRepository
}

// NewBookService создаёт новый сервис
func NewBookService(repo *booktags.BookRepository) *BookService {
	return &BookService{repo: repo}
}

// GetAllBooks возвращает все книги
func (s *BookService) GetAllBooks(ctx context.Context) ([]booktags.BookWithTags, error) {
	// Пока просто вызываем репозиторий
	// В будущем здесь может быть кеширование, фильтрация и т.д.
	return s.repo.GetAllBooks(ctx)
}

func (s *BookService) AddBook(ctx context.Context, bt booktags.BookWithTags) error {
	// Валидация данных книги ( ПОЗЖЕ :) )

	return s.repo.AddBook(ctx, bt)
}

func (s *BookService) BookWithID(ctx context.Context, id int) (booktags.BookWithTags, error) {
	//WARNING
	return s.repo.BookWithID(ctx, id)
}

func (s *BookService) GetFileBookWithID(ctx context.Context, id int) (string, error) {
	return s.repo.GetFileBookWithID(ctx, id)
}

func (s *BookService) AddDownloadCountWithID(ctx context.Context, id int) error {
	return s.repo.AddDownloadCountWithID(ctx, id)
}

func (s *BookService) ShowBooksWithTag(ctx context.Context, tag string) ([]booktags.BookWithTags, error) {
	return s.repo.ShowBooksWithTag(ctx,tag)
}

func (s *BookService) SearchBooksWithTitleDesc(ctx context.Context,query string) ([]booktags.BookWithTags, error) {
	return s.repo.SearchBooksWithTitleDesc(ctx,query)
}


