package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	filepb "github.com/dennisdijkstra/memprint/proto/file"
	"github.com/dennisdijkstra/memprint/shared/events"
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

	mq, err := connectRabbitMQ(os.Getenv("RABBITMQ_URL"))
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

func (s *FileServer) UploadFile(ctx context.Context, req *filepb.UploadFileRequest) (*filepb.UploadFileResponse, error) {
	log.Printf("received upload: user=%s, filename=%s, size=%d bytes", req.UserId, req.Filename, len(req.Content))

	fileID := fmt.Sprintf("file_%d", time.Now().UnixNano())

	tmpPath := fmt.Sprintf("/tmp/%s.bin", fileID)
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open tmp file: %w", err)
	}

	meta := captureMetadata(f.Fd())

	if _, err = f.Write(req.Content); err != nil {
		return nil, fmt.Errorf("write tmp file: %w", err)
	}
	f.Sync()
	f.Close()
	os.Remove(tmpPath)

	log.Printf("metadata captured: %s", meta)

	_, err = s.db.Exec(ctx, `
		INSERT INTO files (id, user_id, filename, status)
        VALUES ($1, $2, $3, 'uploaded')
	`, fileID, req.UserId, req.Filename)
	if err != nil {
		return nil, fmt.Errorf("insert file: %w", err)
	}

	log.Printf("saved file: id=%s, user=%s, filename=%s", fileID, req.UserId, req.Filename)

	_, err = s.db.Exec(ctx, `
		INSERT INTO mem_metadata (
			file_id, pid, tid, heap_addr, heap_size,
			stack_offset, fd, nr_mmap, nr_write,
			nr_fsync, nr_openat, captured_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12
		)
	`, fileID,
		meta.PID, meta.TID,
		meta.HeapAddr, meta.HeapSize,
		meta.StackOffset, meta.FD,
		meta.NRMmap, meta.NRWrite,
		meta.NRFsync, meta.NROpenat,
		meta.CapturedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert metadata: %w", err)
	}

	log.Printf("saved metadata for file: %s", fileID)

	event := events.FileUploadedEvent{
		FileID:    fileID,
		UserID:    req.UserId,
		Filename:  req.Filename,
		Meta:      meta,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	if err := s.mq.publish(ctx, event); err != nil {
		return nil, fmt.Errorf("publish event: %w", err)
	}

	log.Printf("published file.uploaded event for %s", fileID)

	return &filepb.UploadFileResponse{
		FileId: fileID,
		Status: "uploaded",
	}, nil
}
