package audit

import "context"

type Observer interface {
	Handle(ctx context.Context, event Event) error
}
