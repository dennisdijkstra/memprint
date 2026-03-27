package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	rabbitMQURL := os.Getenv("RABBITMQ_URL")

	mq, err := connectRabbitMQ(rabbitMQURL)
	if err != nil {
		log.Fatalf("connect rabbitmq: %v", err)
	}
	defer mq.close()

	storage, err := newStorage(context.Background())
	if err != nil {
		log.Fatalf("connect storage: %v", err)
	}

	handler := &RenderHandler{mq: mq, storage: storage}

	if err := mq.consume(handler.handleFileUploaded); err != nil {
		log.Fatalf("consume: %v", err)
	}

	log.Println("render service waiting for file.uploaded events...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("render service shutting down")
}
