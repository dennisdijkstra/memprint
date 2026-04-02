package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dennisdijkstra/memprint/shared/events"
)

type RenderHandler struct {
	mq       *RabbitMQ
	storage  *Storage
	renderer *RendererClient
}

func (h *RenderHandler) handleFileUploaded(body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var event events.FileUploadedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	log.Printf("render job started: file=%s pid=%d heap=%s",
		event.FileID, event.Meta.PID, event.Meta.HeapAddrHex)

	// call Node renderer via gRPC
	pngBytes, err := h.renderer.render(ctx, event.Meta)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	log.Printf("poster rendered: %dB", len(pngBytes))

	// save to tmp
	outputPath := fmt.Sprintf("/tmp/poster_%s.png", event.FileID)
	if err := os.WriteFile(outputPath, pngBytes, 0600); err != nil {
		return fmt.Errorf("write poster: %w", err)
	}

	// upload to S3
	posterURL, err := h.storage.uploadPoster(ctx, event.FileID, outputPath)
	if err != nil {
		return fmt.Errorf("upload poster: %w", err)
	}

	if err := os.Remove(outputPath); err != nil {
		log.Printf("remove temp file: %v", err)
	}

	log.Printf("poster uploaded: %s", posterURL)

	// publish poster.ready
	posterEvent := events.PosterReadyEvent{
		FileID:    event.FileID,
		UserID:    event.UserID,
		JobID:     fmt.Sprintf("job_%d", time.Now().UnixNano()),
		PosterURL: posterURL,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	if err := h.mq.publish(context.Background(), events.QueuePosterReady, posterEvent); err != nil {
		return fmt.Errorf("publish poster.ready: %w", err)
	}

	log.Printf("published poster.ready for %s", event.FileID)
	return nil
}
