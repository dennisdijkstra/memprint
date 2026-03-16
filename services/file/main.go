package main

import (
	"context"
	"log"
	"net"

	filepb "github.com/dennisdijkstra/memprint/proto/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type FileServer struct {
	filepb.UnimplementedFileServiceServer
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	srv := grpc.NewServer()
	reflection.Register(srv)
	filepb.RegisterFileServiceServer(srv, &FileServer{})

	log.Println("File service is running on port 50051...")
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
