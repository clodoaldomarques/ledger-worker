package ports

import (
	"context"

	domain "github.com/clodoaldomarques/ledger-worker/internal/core/ledger/domain"
)

// ProcessEventUseCase é a porta de entrada acionada pelos adaptadores inbound.
type ProcessEventUseCase interface {
	CreateEvent(ctx context.Context, e domain.Event) error
}
