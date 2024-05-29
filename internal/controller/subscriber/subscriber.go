package subscriber

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
)

type Deps struct {
	Log    *logrus.Logger
	Queue  <-chan service.OrderEvent
	Handle func(context.Context, service.OrderEvent) error
}

func New(d *Deps) *Subscriber {
	return &Subscriber{queue: d.Queue, handle: d.Handle}
}

type Subscriber struct {
	log    *logrus.Logger
	queue  <-chan service.OrderEvent
	handle func(context.Context, service.OrderEvent) error
}

func (s *Subscriber) Run(ctx context.Context) error {
LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, opened := <-s.queue:
			if !opened {
				s.log.Errorf("channel closed, breake loop")
				break LOOP
			}

			if err := s.handle(context.Background(), event); err != nil {
				s.log.Errorf("failed to handle event: %s", err)
			}
		}
	}

	return nil
}
