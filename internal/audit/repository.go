package audit

import "context"

// Repository records and queries audit events.
type Repository interface {
	Sink
	ListRecent(ctx context.Context, limit int) ([]Event, error)
}
