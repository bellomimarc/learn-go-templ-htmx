package todos

import (
	"context"
	"errors"
	"testing"
)

func TestTodoServiceCreateTrimsTitle(t *testing.T) {
	repository := &todoRepositoryStub{}
	service := NewTodoService(repository)

	_, err := service.Create(context.Background(), "  Write service test  ")
	if err != nil {
		t.Fatalf("create TODO: %v", err)
	}
	if repository.createdTitle != "Write service test" {
		t.Fatalf("expected trimmed title, got %q", repository.createdTitle)
	}
}

func TestTodoServiceRejectsBlankTitle(t *testing.T) {
	repository := &todoRepositoryStub{}
	service := NewTodoService(repository)

	_, err := service.Create(context.Background(), "   ")
	if !errors.Is(err, ErrInvalidTitle) {
		t.Fatalf("expected ErrInvalidTitle, got %v", err)
	}
	if repository.createCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repository.createCalls)
	}
}

type todoRepositoryStub struct {
	createdTitle string
	createCalls  int
}

func (repository *todoRepositoryStub) List(context.Context) ([]Todo, error) {
	return nil, nil
}

func (repository *todoRepositoryStub) Create(_ context.Context, title string) (Todo, error) {
	repository.createdTitle = title
	repository.createCalls++
	return Todo{Title: title}, nil
}

func (repository *todoRepositoryStub) Rename(context.Context, int64, string) (Todo, error) {
	return Todo{}, nil
}

func (repository *todoRepositoryStub) Toggle(context.Context, int64) (Todo, error) {
	return Todo{}, nil
}

func (repository *todoRepositoryStub) Delete(context.Context, int64) error {
	return nil
}
