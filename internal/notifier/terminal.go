package notifier

import (
	"fmt"

	"github.com/Trungsherlock/jobgocli/internal/database"
)

type TerminalNotifier struct{}

func NewTerminalNotifier() *TerminalNotifier {
	return &TerminalNotifier{}
}

func (t *TerminalNotifier) Notify(job database.Job, companyName string, score float64) error {
	location := ""
	if job.Location != nil {
		location = *job.Location
	}
	fmt.Printf("	NEW MATCH [%.0f] %s @ %s (%s)\n", score, job.Title, companyName, location)
	fmt.Printf("	-> %s\n", job.URL)
	return nil
}