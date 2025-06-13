package services

import (
	"PropertyAppBackend/config"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// SendSMS sends an SMS message using Twilio Programmable SMS API
func SendSMS(toPhoneNumber, messageBody string) error {
	cfg := config.LoadConfig()
	if cfg.TwilioAccountSID == "" || cfg.TwilioAuthToken == "" || cfg.TwilioPhoneNumber == "" {
		return fmt.Errorf("Twilio SMS credentials not configured")
	}

	urlStr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", cfg.TwilioAccountSID)

	data := url.Values{}
	data.Set("To", toPhoneNumber)
	data.Set("From", cfg.TwilioPhoneNumber) // Your Twilio phone number
	data.Set("Body", messageBody)

	client := &http.Client{}
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create Twilio SMS request: %w", err)
	}

	req.SetBasicAuth(cfg.TwilioAccountSID, cfg.TwilioAuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Twilio SMS request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil // Success
	}

	// Read and log error response from Twilio for debugging
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("Twilio SMS Error Response (%d): %s\n", resp.StatusCode, string(bodyBytes))

	return fmt.Errorf("Twilio SMS API returned status %d for %s", resp.StatusCode, toPhoneNumber)
}