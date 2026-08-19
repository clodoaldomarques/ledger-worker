package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/clodoaldomarques/core-sdk/pkg/logger"
	"github.com/clodoaldomarques/ledger-worker/config"
	"github.com/clodoaldomarques/ledger-worker/internal/domain/ledger"
	"github.com/sony/gobreaker"
)

type LedgerEventsApi struct {
	baseUrl        string
	httpClient     *http.Client
	circuitBreaker *gobreaker.CircuitBreaker
}

func New(ctx context.Context) *LedgerEventsApi {
	baseUrl := config.New().LedgerEventsApiUrl
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
		httpClient:     &http.Client{},
		circuitBreaker: cb,
	}
}

func (a *LedgerEventsApi) CreateEvent(ctx context.Context, e ledger.Event) error {
	_, err := a.circuitBreaker.Execute(func() (interface{}, error) {
		u := fmt.Sprintf("%s/v1/ledger/events", a.baseUrl)

		b, err := json.Marshal(NewEventRequest(e))
		if err != nil {
			return nil, fmt.Errorf("failed to marshal event: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(b))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-cid", e.Cid)
		req.Header.Set("x-tenant", e.OrgID)

		resp, err := a.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusCreated {
			return nil, fmt.Errorf("api error: status %d, body: %s", resp.StatusCode, string(respBody))
		}

		return nil, nil
	})

	return err
}

func (a *LedgerEventsApi) Close() {
	a.httpClient.CloseIdleConnections()
}
