package audit

import "context"

// Observer описывает получателя событий аудита.
type Observer interface {
	// Handle обрабатывает событие аудита event.
	Handle(ctx context.Context, event Event) error
}
