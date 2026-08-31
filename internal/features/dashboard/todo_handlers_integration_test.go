package dashboard

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTodoCRUDLifecycle(t *testing.T) {
	pool, router := setupTodoIntegrationTest(t)

	response := todoRequest(t, router, http.MethodPost, "/dashboard/todos", url.Values{"title": {"  Write integration test  "}})
	assertStatus(t, response, http.StatusCreated)
	assertBodyContains(t, response, "Write integration test")

	var todo Todo
	err := pool.QueryRow(context.Background(), `
		SELECT id, title, completed, created_at, updated_at
		FROM todos`).Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		t.Fatalf("query created TODO: %v", err)
	}
	if todo.Title != "Write integration test" || todo.Completed {
		t.Fatalf("unexpected created TODO: %+v", todo)
	}

	response = todoRequest(t, router, http.MethodGet, "/dashboard/todos", nil)
	assertStatus(t, response, http.StatusOK)
	assertBodyContains(t, response, "Write integration test")

	todoPath := "/dashboard/todos/" + strconv.FormatInt(todo.ID, 10)
	response = todoRequest(t, router, http.MethodPut, todoPath, url.Values{"title": {"Ship integration test"}})
	assertStatus(t, response, http.StatusOK)
	assertBodyContains(t, response, "Ship integration test")
	assertTodoState(t, pool, todo.ID, "Ship integration test", false)

	response = todoRequest(t, router, http.MethodPatch, todoPath+"/toggle", nil)
	assertStatus(t, response, http.StatusOK)
	assertTodoState(t, pool, todo.ID, "Ship integration test", true)

	response = todoRequest(t, router, http.MethodDelete, todoPath, nil)
	assertStatus(t, response, http.StatusOK)
	assertBodyContains(t, response, "Nothing here yet")

	var count int
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM todos`).Scan(&count); err != nil {
		t.Fatalf("count TODOs after delete: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no TODOs after delete, got %d", count)
	}
}

func TestTodoValidationAndMissingRecords(t *testing.T) {
	_, router := setupTodoIntegrationTest(t)

	response := todoRequest(t, router, http.MethodPost, "/dashboard/todos?lang=it", url.Values{"title": {"   "}})
	assertStatus(t, response, http.StatusUnprocessableEntity)
	assertBodyContains(t, response, "Inserisci un titolo")

	response = todoRequest(t, router, http.MethodPut, "/dashboard/todos/999999", url.Values{"title": {"Missing"}})
	assertStatus(t, response, http.StatusNotFound)

	response = todoRequest(t, router, http.MethodPatch, "/dashboard/todos/not-a-number/toggle", nil)
	assertStatus(t, response, http.StatusBadRequest)
}

func setupTodoIntegrationTest(t *testing.T) (*pgxpool.Pool, http.Handler) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("create test database pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping test database: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `TRUNCATE todos RESTART IDENTITY`); err != nil {
		t.Fatalf("reset TODO table: %v", err)
	}

	router := chi.NewRouter()
	RegisterRoutes(router, NewTodoStore(pool))
	return pool, router
}

func todoRequest(t *testing.T, handler http.Handler, method, target string, values url.Values) *http.Response {
	t.Helper()
	var body io.Reader
	if values != nil {
		body = strings.NewReader(values.Encode())
	}
	request := httptest.NewRequest(method, target, body)
	if values != nil {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder.Result()
}

func assertStatus(t *testing.T, response *http.Response, expected int) {
	t.Helper()
	if response.StatusCode != expected {
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected status %d, got %d: %s", expected, response.StatusCode, body)
	}
}

func assertBodyContains(t *testing.T, response *http.Response, expected string) {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if !strings.Contains(string(body), expected) {
		t.Fatalf("expected response body to contain %q, got %s", expected, body)
	}
}

func assertTodoState(t *testing.T, pool *pgxpool.Pool, id int64, title string, completed bool) {
	t.Helper()
	var actualTitle string
	var actualCompleted bool
	if err := pool.QueryRow(context.Background(), `
		SELECT title, completed
		FROM todos
		WHERE id = $1`, id).Scan(&actualTitle, &actualCompleted); err != nil {
		t.Fatalf("query TODO state: %v", err)
	}
	if actualTitle != title || actualCompleted != completed {
		t.Fatalf("expected title=%q completed=%t, got title=%q completed=%t", title, completed, actualTitle, actualCompleted)
	}
}
