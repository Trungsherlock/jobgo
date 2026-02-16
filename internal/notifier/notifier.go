package notifier

import "github.com/Trungsherlock/jobgocli/internal/database"

type Notifier interface {
	Notify(job database.Job, companyName string, score float64) error
}