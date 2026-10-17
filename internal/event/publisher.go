package event

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EventPublisher struct {
	channel *amqp.Channel
}

func NewEventPublisher(ch *amqp.Channel) *EventPublisher {
	return &EventPublisher{channel: ch}
}

func (p *EventPublisher) PublishDomainEvent(ctx context.Context, exchange, routingKey string, event DomainEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}
