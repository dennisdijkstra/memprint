package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dennisdijkstra/memprint/shared/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func connectRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		events.QueueFileUploaded,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	log.Println("connected to rabbitmq")
	return &RabbitMQ{conn: conn, channel: ch}, nil
}

func (r *RabbitMQ) consume(handler func([]byte) error) error {
	msgs, err := r.channel.Consume(
		events.QueueFileUploaded,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("handler error: %v - requeueing", err)
				if err := msg.Nack(false, true); err != nil {
					log.Printf("nack failed: %v", err)
				}
			} else {
				if err := msg.Ack(false); err != nil {
					log.Printf("ack failed: %v", err)
				}
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) close() {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("close channel: %v", err)
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("close connection: %v", err)
		}
	}
}

func (r *RabbitMQ) publish(ctx context.Context, queue string, payload any) error {
	_, err := r.channel.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	return r.channel.PublishWithContext(ctx,
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
