package main

import (
	"context"
	"log"
	"net"
	"os"

	filepb "github.com/dennisdijkstra/memprint/proto/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type FileServer struct {
	filepb.UnimplementedFileServiceServer
	db *pgxpool.Pool
	mq *RabbitMQ
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	port := os.Getenv("FILE_SERVICE_PORT")
	dbURL := os.Getenv("DATABASE_URL")
	rabbitMQURL := os.Getenv("RABBITMQ_URL")

	db, err := connectDB(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer db.Close()
	log.Println("connnected to postgres")

	if err := runMigrations(context.Background(), dbURL); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
	log.Println("migrations applied")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	srv := grpc.NewServer()
	reflection.Register(srv)

	mq, err := connectRabbitMQ(rabbitMQURL)
	if err != nil {
		log.Fatalf("connect rabbitmq: %v", err)
	}
	defer mq.close()

	filepb.RegisterFileServiceServer(srv, &FileServer{
		db: db,
		mq: mq,
	})

	log.Printf("file service listening on :%s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatal("failed to serve: ", err)
	}
}
