package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var notificationBuffer []string

func notifyDiscord(msg string) {
	notificationBuffer = append(notificationBuffer, msg)
}

type Payload struct {
	Content string `json:"content"`
}

func PostToDiscordWebhook(webhookURL string, message string) error {
	data := Payload{
		Content: message,
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", webhookURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("received a status code %d", resp.StatusCode)
	}

	return nil
}

func flushDiscordNotifications() {
	settings, err := loadSettings()
	if err != nil {
		fmt.Println("Error loading settings:", err)
		return
	}

	if settings.EnableDiscordNotifications {
		// Pull contact info from settings file.
		notifyDiscord("===================================== Submission Info =====================================")
		notifyDiscord("Submitted by: " + settings.UserName)
		notifyDiscord("Discord: " + settings.Discord)
		notifyDiscord("Twitter: " + settings.Twitter)
		notifyDiscord("Reddit: " + settings.Reddit)
		notifyDiscord("===================================== End of Report =====================================")

		allMessages := strings.Join(notificationBuffer, "\n")
		err := PostToDiscordWebhook("https://discord.com/api/webhooks/1109543986944811130/J-0O0WyE9aOuyM2XX2fNnGUhbsOCEOXAWSXmzYSNe0S5NNXLEAPZGOwsfNWVO9HI2X5X", allMessages)
		if err != nil {
			fmt.Println("Error posting to Discord:", err)
		}
	}
	notificationBuffer = []string{} // clear the buffer after posting
}