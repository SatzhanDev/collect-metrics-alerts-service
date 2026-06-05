package audit

import (
	"context"

	"go.uber.org/zap"
)

type Publisher struct {
	observers []Observer
	logger    *zap.Logger
}

func NewPublisher(logger *zap.Logger) *Publisher {
	return &Publisher{
		logger: logger,
	}
}

func (p *Publisher) Subscribe(observer Observer) {
	p.observers = append(p.observers, observer)
}

func (p *Publisher) Notify(ctx context.Context, event Event) {
	for _, observer := range p.observers {
		if err := observer.Handle(ctx, event); err != nil {
			p.logger.Error("audit observer failed", zap.Error(err))
		}
	}
}

func (p *Publisher) Enabled() bool {
	return len(p.observers) > 0
}
