package args

import (
	"time"

	"github.com/riverqueue/river"
)

// SendAlertArgs is queued when a service goes down or comes back up. The worker finds who gets
// alert emails for the service and queues one SendEmailArgs per person.
type SendAlertArgs struct {
	ServiceId           string    `json:"service_id"`
	OrgId               string    `json:"org_id"`
	Status              string    `json:"status"` // up or down
	CheckedAt           time.Time `json:"checked_at"`
	StatusCode          *int      `json:"status_code"`
	Error               *string   `json:"error"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
}

func (SendAlertArgs) Kind() string {
	return "send_alert"
}

// InsertOpts: retried a few times if the database is briefly unavailable. The emails are
// queued in one insert, so a retry cannot send some of them twice.
func (SendAlertArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{MaxAttempts: 3}
}

// SendEmailArgs sends one email. Each recipient has their own job, so a retry for one person
// does not send the email again to the others.
type SendEmailArgs struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (SendEmailArgs) Kind() string {
	return "send_email"
}

// InsertOpts: retried with backoff when the mail server is unavailable
func (SendEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{MaxAttempts: 5}
}
