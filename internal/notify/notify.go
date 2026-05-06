// Package notify provides webhook and alerting support for vaultpatch operations.
// It allows callers to dispatch structured event notifications to external
// endpoints (e.g. Slack, PagerDuty, generic webhooks) when secrets are
// changed, drifted, or rotated.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Event represents a vaultpatch lifecycle notification.
type Event struct {
	// Operation is the vaultpatch command that triggered the event
	// (e.g. "apply", "rotate", "drift", "rollback").
	Operation string `json:"operation"`

	// Environment is an optional label describing the target environment.
	Environment string `json:"environment,omitempty"`

	// Paths lists the Vault secret paths affected by the operation.
	Paths []string `json:"paths,omitempty"`

	// DryRun indicates whether the operation was a dry run (no writes).
	DryRun bool `json:"dry_run"`

	// Success reports whether the operation completed without errors.
	Success bool `json:"success"`

	// Message is an optional human-readable summary.
	Message string `json:"message,omitempty"`

	// Timestamp is the UTC time the event was generated.
	Timestamp time.Time `json:"timestamp"`
}

// Config holds the configuration required to dispatch notifications.
type Config struct {
	// WebhookURL is the HTTP(S) endpoint that receives POST requests.
	WebhookURL string

	// Timeout is the per-request deadline. Defaults to 10 s when zero.
	Timeout time.Duration

	// Headers contains optional HTTP headers added to every request
	// (e.g. Authorization tokens).
	Headers map[string]string
}

// Notifier dispatches Event payloads to a configured webhook endpoint.
type Notifier struct {
	cfg    Config
	client *http.Client
}

// New creates a Notifier from cfg. It returns an error if WebhookURL is empty.
func New(cfg Config) (*Notifier, error) {
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("notify: WebhookURL must not be empty")
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Notifier{
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
	}, nil
}

// Send marshals ev as JSON and POSTs it to the configured webhook URL.
// A non-2xx HTTP status is treated as an error.
func (n *Notifier) Send(ctx context.Context, ev Event) error {
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}

	body, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("notify: marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.cfg.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range n.cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("notify: send request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notify: webhook returned HTTP %d", resp.StatusCode)
	}
	return nil
}
