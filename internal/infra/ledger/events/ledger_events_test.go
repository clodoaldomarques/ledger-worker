package events

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/clodoaldomarques/ledger-worker/config"
	"github.com/clodoaldomarques/ledger-worker/internal/domain/ledger"
	"github.com/golang/mock/gomock"
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

		req, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		sr := string(req)

		if !strings.Contains(sr, "error") {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "bad request"}`))
		}
	}))
	defer server.Close()

	type args struct {
		event ledger.Event
	}
	tests := []struct {
		name  string
		setup func(ctrl *gomock.Controller) *LedgerEventsApi
		args  func() args
		want  func(t *testing.T, e error)
	}{
		{
			name: "success - create ledger events",
			setup: func(ctrl *gomock.Controller) *LedgerEventsApi {
				config.New(config.WithLedgerEventsApiUrl(server.URL))
				return New(context.Background())
			},
			args: func() args {
				return args{
					event: fakeEvent(false),
				}
			},
			want: func(t *testing.T, e error) {
				assert.Nil(t, e)
			},
		},
		{
			name: "error - create ledger events",
			setup: func(ctrl *gomock.Controller) *LedgerEventsApi {
				config.New(config.WithLedgerEventsApiUrl(server.URL))
				return New(context.Background())
			},
			args: func() args {
				return args{
					event: fakeEvent(true),
				}
			},
			want: func(t *testing.T, e error) {
				assert.NotNil(t, e)
				assert.Equal(t, "ledger config not found", e.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			api := tt.setup(ctrl)
			err := api.CreateEvent(context.Background(), tt.args().event)
			tt.want(t, err)
		})
	}
}

func fakeEvent(e bool) ledger.Event {
	cid := uuid.NewString()
	if e {
		cid = "error"
	}
	return ledger.Event{
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
