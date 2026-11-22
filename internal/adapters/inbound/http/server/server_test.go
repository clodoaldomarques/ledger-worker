package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clodoaldomarques/ledger-worker/internal/core/ledger/ports"
	"github.com/stretchr/testify/assert"
)

type fakeHealthChecker struct {
	err error
}

var _ ports.HealthChecker = (*fakeHealthChecker)(nil)

func (f *fakeHealthChecker) Health() error { return f.err }

func TestHealthEndpointsUseInjectedChecker(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		health error
		want   int
	}{
		{name: "liveness", path: "/health/live", want: http.StatusOK},
		{name: "readiness healthy", path: "/health/ready", want: http.StatusOK},
		{name: "readiness unhealthy", path: "/health/ready", health: errors.New("unavailable"), want: http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(&fakeHealthChecker{err: tt.health})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)

			s.http.ServeHTTP(rec, req)
			assert.Equal(t, tt.want, rec.Code)
		})
	}
}
