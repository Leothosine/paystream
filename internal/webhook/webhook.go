// Package webhook delivers signed event notifications to registered endpoints.
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Event is a notification emitted for a payment status change.
type Event struct {
	Type      string          `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	Data      json.RawMessage `json:"data"`
}

// Endpoint is a registered webhook destination.
type Endpoint struct {
	URL    string
	Secret string
}

// Notifier delivers events to endpoints, signing each payload with
// HMAC-SHA256 and retrying transient failures with exponential backoff.
type Notifier struct {
	HTTPClient *http.Client
	MaxRetries int
	BaseDelay  time.Duration
}

// NewNotifier returns a Notifier with sensible defaults.
func NewNotifier() *Notifier {
	return &Notifier{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		MaxRetries: 3,
		BaseDelay:  time.Second,
	}
}

// Sign returns the hex-encoded HMAC-SHA256 signature for payload using secret.
func Sign(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// Send delivers an event to the endpoint, retrying non-2xx responses and
// transport errors up to MaxRetries times with exponential backoff.
func (n *Notifier) Send(ctx context.Context, ep Endpoint, ev Event) error {
	payload, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("webhook: marshal event: %w", err)
	}

	signature := Sign(ep.Secret, payload)

	var lastErr error
	for attempt := 0; attempt <= n.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := n.BaseDelay * time.Duration(1<<(attempt-1))
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep.URL, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("webhook: build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-PayStream-Signature", signature)
		req.Header.Set("X-PayStream-Event", ev.Type)

		resp, err := n.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("webhook: delivery attempt %d: %w", attempt+1, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("webhook: delivery attempt %d: endpoint returned status %d", attempt+1, resp.StatusCode)
	}

	return lastErr
}
