package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSEBrokerPublishesToSubscribers(t *testing.T) {
	broker := newSSEBroker()
	events, unsubscribe := broker.subscribe()
	defer unsubscribe()

	published := broker.publish("clicked")

	select {
	case received := <-events:
		if received.ID != published.ID || received.Message != "clicked" {
			t.Fatalf("unexpected event: %+v", received)
		}
	case <-time.After(time.Second):
		t.Fatal("expected subscriber to receive event")
	}
}

func TestSSEStreamStopsOnBrokerShutdown(t *testing.T) {
	broker := newSSEBroker()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/dashboard/events/stream", nil)

	finished := make(chan struct{})
	go func() {
		handleEventsStream(broker)(recorder, request)
		close(finished)
	}()

	broker.shutdown()

	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("expected stream handler to return after broker shutdown")
	}
}

func TestWriteSSEEvent(t *testing.T) {
	recorder := httptest.NewRecorder()
	event := sseEvent{ID: 7, Message: "Button clicked", CreatedAt: time.Date(2026, 9, 1, 14, 30, 0, 0, time.UTC)}

	if err := writeSSEEvent(recorder, event); err != nil {
		t.Fatalf("write SSE event: %v", err)
	}

	body := recorder.Body.String()
	for _, expected := range []string{
		"id: 7\n",
		"event: notification\n",
		`"message":"Button clicked"`,
		`"createdAt":"14:30:00"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected body to contain %q, got %s", expected, body)
		}
	}
}
