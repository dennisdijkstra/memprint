package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dennisdijkstra/memprint/shared/events"
)

func handleFileUploaded(body []byte) error {
	var event events.FileUploadedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	log.Printf("render job started: file=%s pid=%d heap=%s",
		event.FileID, event.Meta.PID, event.Meta.HeapAddrHex)

	layout := makeLayout(event.Meta)
	log.Printf("layout built: %d elements seed=%d", len(layout.Elements), layout.Seed)

	for _, el := range layout.Elements {
		log.Printf("  [%s] content=%q size=%.0f rotation=%.1f opacity=%.2f",
			el.Effect, el.Content, el.Size, el.Rotation, el.Opacity)
	}

	log.Printf("render job done: file=%s", event.FileID)
	return nil
}
