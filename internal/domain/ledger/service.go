package ledger

type Service struct {
	api EventsAPI
}

func New(a EventsAPI) *Service {
	return &Service{
		api: a,
	}
}


