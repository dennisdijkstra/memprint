package main

import (
	"fmt"
	"log"

	"github.com/dennisdijkstra/memprint/shared/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

func connectRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare(events.QueuePosterReady, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	log.Println("connected to rabbitmq")
	return &RabbitMQ{conn: conn, channel: ch}, nil
}

func (r *RabbitMQ) consume(queue string, handler func([]byte) error) error {
	msgs, err := r.channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("handler error: %v — requeuing", err)
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
