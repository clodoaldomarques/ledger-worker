package ledger

import "context"

type Service struct {
	api EventsAPI
}

func New(a EventsAPI) *Service {
	return &Service{
		api: a,
	}
}

func (s Service) CreateEvent(ctx context.Context, e Event) error {
	if err := s.api.CreateEvent(ctx, e); err != nil {
		return err
	}
	return nil
}
