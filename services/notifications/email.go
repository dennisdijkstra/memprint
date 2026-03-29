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
        <div style="font-family: Helvetica Neue, Helvetica, Arial, sans-serif; background: #ffffff; color: #000000; padding: 12px 12px; max-width: 560px;">
            <p style="font-size: 24px; font-weight: 700; margin: 0 0 36px 0;">Memprint</p>
            <h1 style="font-size: 60px; font-weight: 900; line-height: 0.85; letter-spacing: -0.03em; margin: 0 0 4px 0; max-width: 360px">Your poster is ready.</h1>
            <p style="font-size: 13px; margin: 12px 0 36px 0;">%s</p>
            <a href="%s" style="font-size: 24px; font-weight: 800; color: #000000; text-decoration: underline; text-underline-offset: 4px; letter-spacing: -0.01em;">Download poster</a>
            <p style="font-size: 10px; margin: 4px 0 0 0; letter-spacing: 0.05em;">%s</p>
        </div>
    `, event.Timestamp, event.PosterURL, event.FileID)
}
