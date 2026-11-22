package ledger

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/ledger-worker/internal/core/ledger/domain"
	ports "github.com/clodoaldomarques/ledger-worker/internal/core/ledger/ports"
)

var _ ports.ProcessEventUseCase = (*Service)(nil)

type Service struct {
	api ports.LedgerEventPublisher
}

func New(a ports.LedgerEventPublisher) *Service {
	return &Service{
		api: a,
	}
}

func (s Service) CreateEvent(ctx context.Context, e ledger.Event) error {
	span, ctx := tracer.NewSpanFromContext(ctx, "Service::Handler", map[string]any{
		"cid":             e.Cid,
		"org_id":          e.OrgID,
		"program_id":      e.ProgramID,
		"account_id":      e.AccountID,
		"processing_code": e.ProcessingCode,
	})

	defer span.End()

	if err := s.api.CreateEvent(ctx, e); err != nil {
		span.AddEvent(err.Error(), map[string]any{
			"Cid":            e.Cid,
			"OrgID":          e.OrgID,
			"ProgramID":      e.ProgramID,
			"AccountID":      e.AccountID,
			"ProcessingCode": e.ProcessingCode,
		})
		span.SetError(err)
		return err
	}
	return nil
}
