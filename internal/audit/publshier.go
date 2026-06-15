package audit

import (
	"context"

	"go.uber.org/zap"
)

// Publisher рассылает события аудита всем подписанным наблюдателям.
type Publisher struct {
	observers []Observer
	logger    *zap.Logger
}

// NewPublisher создаёт новый Publisher с указанным логгером.
func NewPublisher(logger *zap.Logger) *Publisher {
	return &Publisher{
		logger: logger,
	}
}

// Subscribe добавляет наблюдателя observer в список получателей событий.
func (p *Publisher) Subscribe(observer Observer) {
	p.observers = append(p.observers, observer)
}

// Notify отправляет event всем подписанным наблюдателям.
func (p *Publisher) Notify(ctx context.Context, event Event) {
	for _, observer := range p.observers {
		if err := observer.Handle(ctx, event); err != nil {
			p.logger.Error("audit observer failed", zap.Error(err))
		}
	}
}

// Enabled возвращает true, если есть хотя бы один подписанный наблюдатель.
func (p *Publisher) Enabled() bool {
	return len(p.observers) > 0
}
