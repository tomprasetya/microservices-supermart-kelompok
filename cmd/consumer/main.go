package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tomprasetya/microservices-supermart-kelompok/internal/event"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
if rabbitURL == "" {
    rabbitURL = "amqp://guest:guest@localhost:5672/"
}
conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Gagal koneksi ke RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Gagal membuka channel: %v", err)
	}
	defer ch.Close()

	c := event.NewEventConsumer(ch, nil)

	if err := c.SetupTopology(); err != nil {
		log.Fatalf("Gagal membuat topologi: %v", err)
	}

	c.StartConsuming("inventory.order-placed.queue")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Consumer dihentikan...")
}