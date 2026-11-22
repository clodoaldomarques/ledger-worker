package ledgerapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	domain "github.com/clodoaldomarques/ledger-worker/internal/core/ledger/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestLedgerEventsApi_CreateEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("x-cid"))
		assert.NotEmpty(t, r.Header.Get("x-tenant"))

		w.Header().Set("Content-Type", "application/json")

		if r.Header.Get("x-cid") != "error" {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "bad request"}`))
		}
	}))
	defer server.Close()

	tests := []struct {
		name  string
		event domain.Event
		want  func(t *testing.T, err error)
	}{
		{
			name:  "success - create ledger events",
			event: fakeEvent(false),
			want: func(t *testing.T, e error) {
				assert.Nil(t, e)
			},
		},
		{
			name:  "error - create ledger events",
			event: fakeEvent(true),
			want: func(t *testing.T, e error) {
				assert.NotNil(t, e)
				assert.Equal(t, "api error: status 400, body: {\"message\": \"bad request\"}", e.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := New(context.Background(), server.URL)
			err := api.CreateEvent(context.Background(), tt.event)
			tt.want(t, err)
			api.Close()
		})
	}
}

func fakeEvent(e bool) domain.Event {
	cid := uuid.NewString()
	if e {
		cid = "error"
	}
	return domain.Event{
		Cid:            cid,
		OrgID:          "TN-Teste",
		ProgramID:      123,
		AccountID:      456,
		ProcessingCode: "b-612",
		Amounts: map[string]decimal.Decimal{
			"principal": decimal.NewFromFloat(2.99),
		},
		Fees: map[string]decimal.Decimal{
			"iof": decimal.NewFromFloat(3.99),
		},
	}
}
