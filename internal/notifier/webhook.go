package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Trungsherlock/jobgocli/internal/database"
)

type WebhookNotifier struct {
	url		string
	client	*http.Client
}

func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		url:	url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WebhookNotifier) Notify(job database.Job, companyName string, score float64) error {
	location := ""
	if job.Location != nil {
		location = *job.Location
	}
	payload := map[string]string{
		"text": fmt.Sprintf("*New Job Match [%.0f]*\n*%s* @ %s\nLocation: %s\n<%s|Apply>",
				score, job.Title, companyName, location, job.URL),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling webhook payload: %w", err)
	}

	resp, err := w.client.Post(w.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}