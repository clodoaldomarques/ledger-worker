package ledger

import "context"

type EventsAPI interface {
	CreateEvent(ctx context.Context, e Event) error
}
