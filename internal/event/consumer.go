package event

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EventConsumer struct {
	channel *amqp.Channel
	db      *sql.DB // opsional: set nil jika ingin simulasi in-memory terlebih dahulu
}

func NewEventConsumer(ch *amqp.Channel, db *sql.DB) *EventConsumer {
	return &EventConsumer{channel: ch, db: db}
}

// SetupTopology mendeklarasikan Exchange, Queue, dan DLX sesuai standar AMQP 0-9-1
func (c *EventConsumer) SetupTopology() error {
	// 1. Deklarasi Dead-Letter Exchange (DLX) & Dead-Letter Queue (DLQ)
	err := c.channel.ExchangeDeclare(
		"supermart.dlx", // Nama Exchange
		"direct",         // Tipe Exchange
		true,             // Durable
		false, false, false, nil,
	)
	if err != nil {
		return err
	}

	_, err = c.channel.QueueDeclare(
		"inventory.poison-pill.dlq",
		true, // Durable
		false, false, false, nil,
	)
	if err != nil {
		return err
	}

	err = c.channel.QueueBind(
		"inventory.poison-pill.dlq",
		"inventory.dead-letter",
		"supermart.dlx",
		false, nil,
	)
	if err != nil {
		return err
	}

	// 2. Deklarasi Main Exchange (Tipe Topic)
	err = c.channel.ExchangeDeclare(
		"supermart.order.events",
		"topic",
		true, // Durable
		false, false, false, nil,
	)
	if err != nil {
		return err
	}

	// 3. Deklarasi Main Queue dengan argumen penampung DLX
	args := amqp.Table{
		"x-dead-letter-exchange":    "supermart.dlx",
		"x-dead-letter-routing-key": "inventory.dead-letter",
	}

	_, err = c.channel.QueueDeclare(
		"inventory.order-placed.queue",
		true, // Durable
		false, false, false, args,
	)
	if err != nil {
		return err
	}

	// 4. Binding Main Queue ke Main Exchange
	return c.channel.QueueBind(
		"inventory.order-placed.queue",
		"order.event.*",
		"supermart.order.events",
		false, nil,
	)
}

// StartConsuming mendengarkan pesan masuk secara asinkron
func (c *EventConsumer) StartConsuming(queueName string) {
	// Batasi kuota unacked messages ke worker (mencegah Out-Of-Memory)
	err := c.channel.Qos(10, 0, false)
	if err != nil {
		log.Fatalf("Gagal menyetel QoS: %v", err)
	}

	// Wajib auto_ack = false untuk keandalan pesan
	msgs, err := c.channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack = FALSE
		false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("Gagal mendaftarkan consumer: %v", err)
	}

	log.Printf("[*] Consumer aktif. Menunggu event dari antrean: %s", queueName)

	go func() {
		for d := range msgs {
			err := c.processMessage(d.Body)
			if err != nil {
				log.Printf("[REJECT] Pesan rusak atau invalid: %v -> dialihkan ke DLX", err)
				// d.Reject(false) membuang pesan ke DLX tanpa memasukkannya kembali ke antrean utama
				_ = d.Reject(false)
				continue
			}

			// Manual ACK setelah proses bisnis sukses
			_ = d.Ack(false)
			log.Println("[ACK] Pesan berhasil diproses dan di-commit.")
		}
	}()
}

func (c *EventConsumer) processMessage(body []byte) error {
	var event DomainEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event_id kosong (terindikasi poison pill)")
	}

	log.Printf("[PROCESS] Memproses Event ID: %s | Order ID: %s", event.EventID, event.Data.OrderID)
	return nil
}