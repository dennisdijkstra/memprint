package main

import (
	"log"
	"net/http"
	"os"

	filepb "github.com/dennisdijkstra/memprint/proto/file"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	fileClient filepb.FileServiceClient
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	port := os.Getenv("GATEWAY_PORT")
	fsAddr := os.Getenv("FILE_SERVICE_ADDR")

	conn, err := grpc.Dial(
		fsAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("connect to file service: %v", err)
	}
	defer conn.Close()

	gw := &Gateway{
		fileClient: filepb.NewFileServiceClient(conn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /upload", gw.handleUpload)
	mux.HandleFunc("GET /health", handleHealth)

	log.Printf("gateway listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}

	log.Printf("gateway listening on: %s", port)

}

func (gw *Gateway) handleUpload(w http.ResponseWriter, r *http.Request) {
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
}
