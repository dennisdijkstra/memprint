package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dennisdijkstra/memprint/shared/events"
	"github.com/resend/resend-go/v2"
)

func (h *NotificationHandler) sendEmail(event events.PosterReadyEvent) error {
	params := &resend.SendEmailRequest{
		From:    os.Getenv("RESEND_FROM"),
		To:      []string{os.Getenv("RESEND_TO")},
		Subject: "Your Memprint poster is ready",
		Html:    buildEmailHTML(event),
	}

	resp, err := h.resend.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	log.Printf("email sent id=%s to=%s", resp.Id, os.Getenv("RESEND_TO"))
	return nil
}

func buildEmailHTML(event events.PosterReadyEvent) string {
	return fmt.Sprintf(`
        <div style="font-family: monospace; background: #080808; color: #f0f0f0; padding: 40px;">
            <h1 style="color: #00CFFF; font-size: 24px;">MEMPRINT</h1>
            <p style="color: #888; font-size: 12px;">memory fingerprint generated</p>
            <hr style="border-color: #222; margin: 24px 0"/>
            <p>Your poster is ready.</p>
            <p style="margin: 24px 0">
                <a href="%s"
                   style="background: #00CFFF; color: #080808; padding: 12px 24px;
                          text-decoration: none; font-weight: bold;">
                    Download Poster
                </a>
            </p>
            <hr style="border-color: #222; margin: 24px 0"/>
            <p style="color: #444; font-size: 10px;">file: %s · %s</p>
        </div>
    `, event.PosterURL, event.FileID, event.Timestamp)
}
