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
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}
	log.Println("DATABASE_URL:", os.Getenv("DATABASE_URL"))

	port := os.Getenv("FILE_SERVICE_PORT")
	dbUrl := os.Getenv("DATABASE_URL")

	db, err := connectDB(context.Background(), dbUrl)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer db.Close()
	log.Println("connnected to postgres")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	srv := grpc.NewServer()
	reflection.Register(srv)
	filepb.RegisterFileServiceServer(srv, &FileServer{db: db})

	log.Printf("file service listening on :%s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatal("failed to serve: ", err)
	}
}

func (s *FileServer) UploadFile(ctx context.Context, req *filepb.UploadFileRequest) (*filepb.UploadFileResponse, error) {
	log.Printf("received upload: user=%s, filename=%s, size=%d bytes", req.UserId, req.Filename, len(req.Content))

	return &filepb.UploadFileResponse{
		FileId: "file_001",
		Status: "uploaded",
	}, nil
}
