package todos

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTodoRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTodoRepository(pool *pgxpool.Pool) *PostgresTodoRepository {
	return &PostgresTodoRepository{pool: pool}
}

func (repository *PostgresTodoRepository) List(ctx context.Context) ([]Todo, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT id, title, completed, created_at, updated_at
		FROM todos
		ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[Todo])
	if err != nil {
		return nil, err
	}
	return todos, nil
}

func (repository *PostgresTodoRepository) Create(ctx context.Context, title string) (Todo, error) {
	title, err := validateTitle(title)
	if err != nil {
		return Todo{}, err
	}

	return scanTodo(repository.pool.QueryRow(ctx, `
		INSERT INTO todos (title)
		VALUES ($1)
		RETURNING id, title, completed, created_at, updated_at`, title))
}

func (repository *PostgresTodoRepository) Rename(ctx context.Context, id int64, title string) (Todo, error) {
	title, err := validateTitle(title)
	if err != nil {
		return Todo{}, err
	}

	todo, err := scanTodo(repository.pool.QueryRow(ctx, `
		UPDATE todos
		SET title = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, completed, created_at, updated_at`, id, title))
	return todo, normalizeNotFound(err)
}

func (repository *PostgresTodoRepository) Toggle(ctx context.Context, id int64) (Todo, error) {
	todo, err := scanTodo(repository.pool.QueryRow(ctx, `
		UPDATE todos
		SET completed = NOT completed, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, completed, created_at, updated_at`, id))
	return todo, normalizeNotFound(err)
}

func (repository *PostgresTodoRepository) Delete(ctx context.Context, id int64) error {
	result, err := repository.pool.Exec(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type todoRow interface {
	Scan(dest ...any) error
}

func scanTodo(row todoRow) (Todo, error) {
	var todo Todo
	err := row.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	return todo, err
}

func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", ErrInvalidTitle
	}
	return title, nil
}

func normalizeNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
