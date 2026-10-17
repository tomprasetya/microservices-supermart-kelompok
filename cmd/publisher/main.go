package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tomprasetya/microservices-supermart-kelompok/internal/event"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Gagal koneksi ke RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Gagal membuka channel: %v", err)
	}
	defer ch.Close()

	pub := event.NewEventPublisher(ch)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Mulai menerbitkan 50 event...")

	for i := 1; i <= 50; i++ {
		ev := event.DomainEvent{
			EventID:   uuid.New().String(),
			EventType: "OrderPlaced",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   "1.0",
			Producer:  "order-service",
			Data: event.OrderPlacedPayload{
				OrderID:    fmt.Sprintf("ORD-20261017-%03d", i),
				CustomerID: "CUST-8812",
				Items: []event.Item{
					{ProductID: "prod-001", Quantity: 2},
				},
			},
		}

		err := pub.PublishDomainEvent(ctx, "supermart.order.events", "order.event.placed", ev)
		if err != nil {
			log.Fatalf("Gagal publish pesan ke-%d: %v", i, err)
		}
	}

	log.Println("Selesai menerbitkan 50 event ke RabbitMQ.")
}