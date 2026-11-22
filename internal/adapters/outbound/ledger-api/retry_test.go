package ledgerapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEventRetriesTransientFailure(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	api := newTestAPI(server.URL)
	err := api.CreateEvent(context.Background(), fakeEvent(false))
	require.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls))
}

func TestCreateEventDoesNotRetryClientError(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "bad request"}`))
	}))
	defer server.Close()

	api := newTestAPI(server.URL)
	err := api.CreateEvent(context.Background(), fakeEvent(false))
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

func TestCreateEventCircuitBreakerOpens(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	api := newTestAPI(server.URL)
	for i := 0; i < 6; i++ {
		_ = api.CreateEvent(context.Background(), fakeEvent(false))
	}
	callsAfterSix := atomic.LoadInt32(&calls)
	assert.Equal(t, int32(6*defaultRetryAttempts), callsAfterSix)

	err := api.CreateEvent(context.Background(), fakeEvent(false))
	assert.ErrorIs(t, err, gobreaker.ErrOpenState)
	assert.Equal(t, callsAfterSix, atomic.LoadInt32(&calls))
}

func TestCreateEventSendsIdempotencyKey(t *testing.T) {
	event := fakeEvent(false)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, event.Cid, r.Header.Get("Idempotency-Key"))
		assert.Equal(t, event.Cid, r.Header.Get("x-cid"))
		assert.Equal(t, event.OrgID, r.Header.Get("x-tenant"))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	api := newTestAPI(server.URL)
	require.NoError(t, api.CreateEvent(context.Background(), event))
}

func newTestAPI(baseURL string) *LedgerEventsApi {
	return &LedgerEventsApi{
		baseUrl:    baseURL,
		httpClient: http.DefaultClient,
		circuitBreaker: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "test",
			Timeout: 15 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 5
			},
		}),
		retryAttempts:  defaultRetryAttempts,
		retryBaseDelay: time.Millisecond,
		retryMaxDelay:  2 * time.Millisecond,
	}
}
