package ledger

import (
	"context"

	"github.com/clodoaldomarques/ledger-worker/internal/domain/ledger"
)

//go:generate mockgen -source=interfaces.go -destination=mock.go -package=ledger
type EventsProvider interface {
	CreateEvent(ctx context.Context, e ledger.Event) error
}
