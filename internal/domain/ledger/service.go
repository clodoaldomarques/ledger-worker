package ledger

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/tracer"
	"go.opentelemetry.io/otel/attribute"
)

type Service struct {
	api EventsAPI
}

func New(a EventsAPI) *Service {
	return &Service{
		api: a,
	}
}

func (s Service) CreateEvent(ctx context.Context, e Event) error {
	span, ctx := tracer.NewSpanFromContext(ctx, "Service::Handler", attribute.String("MessageID", e.Cid))
	defer span.End()

	if err := s.api.CreateEvent(ctx, e); err != nil {
		span.AddEvent(err.Error(), tracer.Attributes{
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
