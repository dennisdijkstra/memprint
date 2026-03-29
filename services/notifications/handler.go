package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dennisdijkstra/memprint/shared/events"
)

func (h *NotificationHandler) handlePosterReady(body []byte) error {
	var event events.PosterReadyEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshall event: %w", err)
	}

	log.Printf("poster ready: file=%s user=%s url=%s", event.FileID, event.UserID, event.PosterURL)

	return h.sendEmail(event)
}
