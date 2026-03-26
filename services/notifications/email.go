package main

import "log"

func sendEmail(userID, posterURL string) error {
	// stub - logs for now
	log.Printf("[email stub] to=user:%s subject='Your Memprint poster is ready'", userID)
	log.Printf("[email stub] download: %s", posterURL)
	return nil
}
