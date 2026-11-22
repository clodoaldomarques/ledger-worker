package ledger

import (
	"context"
	"errors"
	"testing"

	domain "github.com/clodoaldomarques/ledger-worker/internal/core/ledger/domain"
	"github.com/clodoaldomarques/ledger-worker/internal/core/ledger/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePublisher struct {
	events []domain.Event
	err    error
}

var _ ports.LedgerEventPublisher = (*fakePublisher)(nil)

func (f *fakePublisher) CreateEvent(_ context.Context, e domain.Event) error {
	f.events = append(f.events, e)
	return f.err
}

// TestService_CreateEventWithFakePublisher executa o mesmo contrato da suíte com
// um fake da porta de saída, sem depender do adaptador HTTP.
func TestService_CreateEventWithFakePublisher(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want func(t *testing.T, err error)
	}{
		{
			name: "success",
			want: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "error",
			err:  errors.New("wrong event"),
			want: func(t *testing.T, err error) {
				assert.EqualError(t, err, "wrong event")
			},
		},
	}

	event := fakeEvent()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := &fakePublisher{err: tt.err}
			svc := New(publisher)

			err := svc.CreateEvent(context.Background(), event)
			tt.want(t, err)
			require.Equal(t, []domain.Event{event}, publisher.events)
		})
	}
}
