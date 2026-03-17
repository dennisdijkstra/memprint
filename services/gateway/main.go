package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

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
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
	}
	defer file.Close()

	buf := make([]byte, header.Size)
	if _, err := file.Read(buf); err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	resp, err := gw.fileClient.UploadFile(ctx, &filepb.UploadFileRequest{
		UserId:   r.FormValue("user_id"),
		Filename: header.Filename,
		Content:  buf,
	})
	if err != nil {
		log.Printf("file service error: %v", err)
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusAccepted, map[string]any{
		"file_id": resp.FileId,
		"status":  resp.Status,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "applicatioin/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
