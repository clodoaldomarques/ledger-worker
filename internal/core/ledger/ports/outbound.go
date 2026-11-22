package ports

import (
	"context"

	domain "github.com/clodoaldomarques/ledger-worker/internal/core/ledger/domain"
)

//go:generate mockgen -destination=mock.go -package=ports github.com/clodoaldomarques/ledger-worker/internal/core/ledger/ports LedgerEventPublisher,HealthChecker,ProcessEventUseCase

// LedgerEventPublisher publica eventos do ledger na API downstream.
type LedgerEventPublisher interface {
	CreateEvent(ctx context.Context, e domain.Event) error
}

// HealthChecker expõe o estado do adaptador de entrada para a probe de readiness.
type HealthChecker interface {
	Health() error
}
