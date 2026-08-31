package dashboard

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	dashboardviews "github.com/marcello/saas-poc/internal/features/dashboard/views"
)

func handleTodoPage(store *TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))
		todos, err := store.List(r.Context())
		if err != nil {
			writeTodoError(w, err)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := dashboardviews.TodoPage(locale, todoViews(todos), "").Render(r.Context(), w); err != nil {
			writeTodoError(w, err)
		}
	}
}

func handleTodoCreate(store *TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := store.Create(r.Context(), r.FormValue("title"))
		if errors.Is(err, ErrInvalidTitle) {
			locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))
			renderTodoRegion(w, r, store, locale.Text("todo.error.title_required"), http.StatusUnprocessableEntity)
			return
		}
		if err != nil {
			writeTodoError(w, err)
			return
		}
		renderTodoRegion(w, r, store, "", http.StatusCreated)
	}
}

func handleTodoRename(store *TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := todoID(w, r)
		if !ok {
			return
		}

		_, err := store.Rename(r.Context(), id, r.FormValue("title"))
		if errors.Is(err, ErrInvalidTitle) {
			locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))
			renderTodoRegion(w, r, store, locale.Text("todo.error.title_required"), http.StatusUnprocessableEntity)
			return
		}
		if !handleTodoMutationError(w, err) {
			return
		}
		renderTodoRegion(w, r, store, "", http.StatusOK)
	}
}

func handleTodoToggle(store *TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := todoID(w, r)
		if !ok {
			return
		}
		_, err := store.Toggle(r.Context(), id)
		if !handleTodoMutationError(w, err) {
			return
		}
		renderTodoRegion(w, r, store, "", http.StatusOK)
	}
}

func handleTodoDelete(store *TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := todoID(w, r)
		if !ok {
			return
		}
		if !handleTodoMutationError(w, store.Delete(r.Context(), id)) {
			return
		}
		renderTodoRegion(w, r, store, "", http.StatusOK)
	}
}

func renderTodoRegion(w http.ResponseWriter, r *http.Request, store *TodoStore, message string, status int) {
	todos, err := store.List(r.Context())
	if err != nil {
		writeTodoError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))
	if err := dashboardviews.TodoRegion(locale, todoViews(todos), message).Render(r.Context(), w); err != nil {
		log.Printf("render TODO region: %v", err)
	}
}

func todoViews(todos []Todo) []dashboardviews.TodoView {
	result := make([]dashboardviews.TodoView, len(todos))
	for index, todo := range todos {
		result[index] = dashboardviews.TodoView{
			ID:        todo.ID,
			Title:     todo.Title,
			Completed: todo.Completed,
		}
	}
	return result
}

func todoID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "invalid TODO id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func handleTodoMutationError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, ErrTodoNotFound) {
		http.Error(w, "TODO not found", http.StatusNotFound)
		return false
	}
	writeTodoError(w, err)
	return false
}

func writeTodoError(w http.ResponseWriter, err error) {
	log.Printf("TODO request failed: %v", err)
	http.Error(w, "Unable to process TODO request", http.StatusInternalServerError)
}
