package ledgerapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/otel/httpclient"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"
	"github.com/clodoaldomarques/ledger-worker/internal/core/ledger/domain"
	"github.com/clodoaldomarques/ledger-worker/internal/core/ledger/ports"
	"github.com/sony/gobreaker"
)

const (
	defaultRetryAttempts  = 3
	defaultRetryBaseDelay = 100 * time.Millisecond
	defaultRetryMaxDelay  = time.Second
)

var _ ports.LedgerEventPublisher = (*LedgerEventsApi)(nil)

type LedgerEventsApi struct {
	baseUrl        string
	httpClient     *http.Client
	circuitBreaker *gobreaker.CircuitBreaker
	retryAttempts  int
	retryBaseDelay time.Duration
	retryMaxDelay  time.Duration
}

func New(ctx context.Context, baseUrl string) *LedgerEventsApi {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:    "LedgerEventsAPI",
		Timeout: 15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Error(ctx, fmt.Sprintf("circuitbreaker '%s' changed from %v to %v", name, from, to), logger.Fields{
				"api_name": name,
				"api_url":  baseUrl,
			})
		},
	})

	return &LedgerEventsApi{
		baseUrl:        baseUrl,
		httpClient:     httpclient.Client(),
		circuitBreaker: cb,
		retryAttempts:  defaultRetryAttempts,
		retryBaseDelay: defaultRetryBaseDelay,
		retryMaxDelay:  defaultRetryMaxDelay,
	}
}

func (a *LedgerEventsApi) CreateEvent(ctx context.Context, e ledger.Event) error {
	span, ctx := tracer.NewSpanFromContext(ctx, "LedgerEventsApi::CreateEvent", map[string]any{
		"cid":        e.Cid,
		"org_id":     e.OrgID,
		"program_id": e.ProgramID,
		"event_type": e,
	})
	defer span.End()

	_, err := a.circuitBreaker.Execute(func() (interface{}, error) {
		u := fmt.Sprintf("%s/v1/ledger/events", a.baseUrl)

		b, err := json.Marshal(NewEventRequest(e))
		if err != nil {
			span.SetError(err)
			return nil, fmt.Errorf("failed to marshal event: %w", err)
		}

		attempts := a.retryAttempts
		if attempts < 1 {
			attempts = 1
		}

		var lastErr error
		for attempt := 1; attempt <= attempts; attempt++ {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
			if err != nil {
				span.SetError(err)
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("x-cid", e.Cid)
			req.Header.Set("x-tenant", e.OrgID)
			req.Header.Set("Idempotency-Key", e.Cid)

			resp, err := a.httpClient.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("http request failed: %w", err)
			} else {
				respBody, readErr := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if readErr != nil {
					lastErr = fmt.Errorf("failed to read response: %w", readErr)
				} else if resp.StatusCode == http.StatusCreated {
					return nil, nil
				} else {
					apiErr := fmt.Errorf("api error: status %d, body: %s", resp.StatusCode, string(respBody))
					if !isRetryable(resp.StatusCode) {
						span.SetError(apiErr)
						return nil, apiErr
					}
					lastErr = apiErr
				}
			}

			if attempt == attempts {
				break
			}

			delay := retryDelay(a.retryBaseDelay, a.retryMaxDelay, attempt)
			span.AddEvent("retrying downstream call", map[string]any{
				"attempt":  attempt + 1,
				"delay_ms": delay.Milliseconds(),
				"error":    lastErr.Error(),
			})
			if err := wait(ctx, delay); err != nil {
				span.SetError(err)
				return nil, err
			}
		}

		span.SetError(lastErr)
		return nil, lastErr
	})

	if err != nil {
		span.SetError(err)
	}
	return err
}

func (a *LedgerEventsApi) Close() {
	a.httpClient.CloseIdleConnections()
}

func isRetryable(status int) bool {
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

func retryDelay(base, max time.Duration, attempt int) time.Duration {
	delay := base * time.Duration(1<<(attempt-1))
	if delay > max {
		delay = max
	}
	delay = time.Duration(float64(delay) * (0.5 + rand.Float64()))
	if delay > max {
		delay = max
	}
	return delay
}

func wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
