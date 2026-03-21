package main

import (
	"context"
	"log"
	"net/http"
	"time"

	filepb "github.com/dennisdijkstra/memprint/proto/file"
)

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
