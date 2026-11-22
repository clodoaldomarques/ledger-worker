package ledger

import (
	"context"

	"github.com/clodoaldomarques/ledger-worker/internal/core/ledger/domain"
)

//go:generate mockgen -source=outbound.go -destination=mock.go -package=ledger
type EventsProvider interface {
	CreateEvent(ctx context.Context, e ledger.Event) error
}
