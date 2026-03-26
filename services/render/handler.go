package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dennisdijkstra/memprint/shared/events"
)

type RenderHandler struct {
	mq *RabbitMQ
}

func (h *RenderHandler) handleFileUploaded(body []byte) error {
	var event events.FileUploadedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	log.Printf("render job started: file=%s pid=%d heap=%s",
		event.FileID, event.Meta.PID, event.Meta.HeapAddrHex)

	layout := makeLayout(event.Meta)
	log.Printf("layout built: %d elements seed=%d", len(layout.Elements), layout.Seed)

	dc, err := renderPoster(layout)
	if err != nil {
		return fmt.Errorf("render poster: %w", err)
	}

	outputPath := fmt.Sprintf("/tmp/poster_%s.png", event.FileID)
	if err := dc.SavePNG(outputPath); err != nil {
		return fmt.Errorf("save poster: %w", err)
	}

	log.Printf("poster saved: %s", outputPath)

	posterEvent := events.PosterReadyEvent{
		FileID:    event.FileID,
		UserID:    event.UserID,
		PosterURL: outputPath,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	if err := h.mq.publish(context.Background(), events.QueuePosterReady, posterEvent); err != nil {
		return fmt.Errorf("publish poster.ready: %w", err)
	}

	log.Printf("published poster.ready for %s", event.FileID)
	return nil
}
