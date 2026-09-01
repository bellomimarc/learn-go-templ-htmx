package dashboard

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	dashboardviews "github.com/marcello/saas-poc/internal/features/dashboard/views"
)

type sseEvent struct {
	ID        int64
	Message   string
	CreatedAt time.Time
}

type sseBroker struct {
	mu          sync.Mutex
	nextID      int64
	subscribers map[chan sseEvent]struct{}
}

const sseHeartbeatInterval = 15 * time.Second

// this is a simple in memory SSE broker implementation, this is not suitable for horizontal scaling.
func newSSEBroker() *sseBroker {
	return &sseBroker{subscribers: map[chan sseEvent]struct{}{}}
}

func (broker *sseBroker) subscribe() (chan sseEvent, func()) {
	events := make(chan sseEvent, 8)

	broker.mu.Lock()
	broker.subscribers[events] = struct{}{}
	broker.mu.Unlock()

	return events, func() {
		broker.mu.Lock()
		delete(broker.subscribers, events)
		close(events)
		broker.mu.Unlock()
	}
}

func (broker *sseBroker) publish(message string) sseEvent {
	broker.mu.Lock()
	broker.nextID++
	event := sseEvent{ID: broker.nextID, Message: message, CreatedAt: time.Now()}
	for subscriber := range broker.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
	broker.mu.Unlock()
	return event
}

func handleEventsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))
	if err := dashboardviews.EventsPage(locale).Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering events page", http.StatusInternalServerError)
		log.Printf("Error rendering events page: %v\n", err)
	}
}

func handleEventTrigger(broker *sseBroker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locale := dashboardviews.LoadLocale(r.URL.Query().Get("lang"))
		message := fmt.Sprintf(locale.Text("events.message.template"), time.Now().Format("15:04:05"))
		broker.publish(message)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleEventsStream(broker *sseBroker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}
		if err := http.NewResponseController(w).SetWriteDeadline(time.Time{}); err != nil {
			log.Printf("disable SSE write deadline: %v", err)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		events, unsubscribe := broker.subscribe()
		defer unsubscribe()

		fmt.Fprint(w, ": connected\n\n")
		flusher.Flush()

		keepAlive := time.NewTicker(sseHeartbeatInterval)
		defer keepAlive.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-keepAlive.C:
				fmt.Fprint(w, ": keep-alive\n\n")
				flusher.Flush()
			case event := <-events:
				if err := writeSSEEvent(w, event); err != nil {
					log.Printf("write SSE event: %v", err)
					return
				}
				flusher.Flush()
			}
		}
	}
}

func writeSSEEvent(w http.ResponseWriter, event sseEvent) error {
	payload := struct {
		ID        int64  `json:"id"`
		Message   string `json:"message"`
		CreatedAt string `json:"createdAt"`
	}{
		ID:        event.ID,
		Message:   event.Message,
		CreatedAt: event.CreatedAt.Format("15:04:05"),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "id: %d\nevent: notification\ndata: %s\n\n", event.ID, data)
	return err
}
