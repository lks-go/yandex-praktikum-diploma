package publisher

import (
	"context"

	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
)

func New() (*Publisher, <-chan service.OrderEvent) {
	ch := make(chan service.OrderEvent, 10)

	return &Publisher{queue: ch}, ch
}

type Publisher struct {
	queue chan<- service.OrderEvent
}

func (p *Publisher) Publish(ctx context.Context, msg service.OrderEvent) {
	p.queue <- msg
}

func (p *Publisher) Close() {
	close(p.queue)
}
