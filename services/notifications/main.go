package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dennisdijkstra/memprint/shared/events"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/resend/resend-go/v2"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type NotificationHandler struct {
	resend *resend.Client
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		log.Fatal("RABBITMQ_URL is not set")
	}

	mq, err := connectRabbitMQ(rabbitMQURL)
	if err != nil {
		log.Fatalf("connect rabbitmq: %v", err)
	}
	defer mq.close()

	resendAPIKey := os.Getenv("RESEND_API_KEY")
	if resendAPIKey == "" {
		log.Fatal("RESEND_API_KEY is not set")
	}

	handler := &NotificationHandler{
		resend: resend.NewClient(os.Getenv("RESEND_API_KEY")),
	}

	if err := mq.consume(events.QueuePosterReady, handler.handlePosterReady); err != nil {
		log.Fatalf("consume: %v", err)
	}

	log.Println("notification service waiting for poster.ready events...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("notification service shutting down")
}
