package todos

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("todo not found")
	ErrInvalidTitle = errors.New("todo title must not be blank")
)

type Todo struct {
	ID        int64
	Title     string
	Completed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TodoRepository interface {
	List(ctx context.Context) ([]Todo, error)
	Create(ctx context.Context, title string) (Todo, error)
	Rename(ctx context.Context, id int64, title string) (Todo, error)
	Toggle(ctx context.Context, id int64) (Todo, error)
	Delete(ctx context.Context, id int64) error
}
