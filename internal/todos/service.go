package todos

import (
	"context"
	"strings"
)

type TodoService interface {
	List(ctx context.Context) ([]Todo, error)
	Create(ctx context.Context, title string) (Todo, error)
	Rename(ctx context.Context, id int64, title string) (Todo, error)
	Toggle(ctx context.Context, id int64) (Todo, error)
	Delete(ctx context.Context, id int64) error
}

type todoService struct {
	repository TodoRepository
}

func NewTodoService(repository TodoRepository) TodoService {
	return &todoService{repository: repository}
}

func (service *todoService) List(ctx context.Context) ([]Todo, error) {
	return service.repository.List(ctx)
}

func (service *todoService) Create(ctx context.Context, title string) (Todo, error) {
	title, err := validateTitle(title)
	if err != nil {
		return Todo{}, err
	}
	return service.repository.Create(ctx, title)
}

func (service *todoService) Rename(ctx context.Context, id int64, title string) (Todo, error) {
	title, err := validateTitle(title)
	if err != nil {
		return Todo{}, err
	}
	return service.repository.Rename(ctx, id, title)
}

func (service *todoService) Toggle(ctx context.Context, id int64) (Todo, error) {
	return service.repository.Toggle(ctx, id)
}

func (service *todoService) Delete(ctx context.Context, id int64) error {
	return service.repository.Delete(ctx, id)
}

func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", ErrInvalidTitle
	}
	return title, nil
}
