package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSign_isDeterministicAndVerifiable(t *testing.T) {
	payload := []byte(`{"type":"payout.completed"}`)

	sig1 := Sign("secret", payload)
	sig2 := Sign("secret", payload)

	if sig1 != sig2 {
		t.Fatalf("expected deterministic signature, got %q and %q", sig1, sig2)
	}

	if Sign("other-secret", payload) == sig1 {
		t.Fatalf("expected different secret to produce a different signature")
	}
}

func TestNotifier_Send_succeedsOnFirstAttempt(t *testing.T) {
	var gotSignature, gotEvent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSignature = r.Header.Get("X-PayStream-Signature")
		gotEvent = r.Header.Get("X-PayStream-Event")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewNotifier()
	n.BaseDelay = time.Millisecond

	ev := Event{Type: "payout.completed", CreatedAt: time.Now(), Data: json.RawMessage(`{"batch_id":"bat_1"}`)}
	ep := Endpoint{URL: srv.URL, Secret: "whsec_test"}

	if err := n.Send(context.Background(), ep, ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotEvent != "payout.completed" {
		t.Fatalf("expected event header payout.completed, got %q", gotEvent)
	}
	if gotSignature == "" {
		t.Fatalf("expected a signature header to be set")
	}
}

func TestNotifier_Send_retriesThenSucceeds(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewNotifier()
	n.BaseDelay = time.Millisecond
	n.MaxRetries = 3

	ev := Event{Type: "payout.failed", CreatedAt: time.Now()}
	ep := Endpoint{URL: srv.URL, Secret: "whsec_test"}

	if err := n.Send(context.Background(), ep, ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestNotifier_Send_returnsErrorAfterExhaustingRetries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := NewNotifier()
	n.BaseDelay = time.Millisecond
	n.MaxRetries = 1

	ev := Event{Type: "payout.failed", CreatedAt: time.Now()}
	ep := Endpoint{URL: srv.URL, Secret: "whsec_test"}

	if err := n.Send(context.Background(), ep, ev); err == nil {
		t.Fatalf("expected an error after exhausting retries")
	}
}
