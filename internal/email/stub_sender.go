package email

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type stubSender struct {
	client HTTPClient
}

func NewStubSender(client HTTPClient) *stubSender {
	return &stubSender{client: client}
}

func (s *stubSender) SendEmail(address string, text string) error {
	payload := map[string]string{
		"to":   address,
		"text": text,
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://example.com/send-email", bytes.NewBuffer(body))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return err
	}

	return nil
}
