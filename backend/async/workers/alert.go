package workers

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"pingwise/async/args"
	"pingwise/async/async_middleware"
	"pingwise/common/custom"
	"pingwise/common/email"
	"pingwise/common/logger"
	"pingwise/internal/models/enums"
	"pingwise/internal/service/notification_service"
)

// sendAlertWorker queues one email for every member who gets alerts for the service
func sendAlertWorker(ctx custom.Context, job *river.Job[args.SendAlertArgs]) error {
	if !email.Enabled() {
		logger.Infof("alert for service %s (%s) not emailed: SMTP is not configured", job.Args.ServiceId, job.Args.Status)
		return nil
	}
	emails, serviceErr := notification_service.GetNotificationService().BuildAlertEmails(ctx, &notification_service.Alert{
		ServiceId:           job.Args.ServiceId,
		OrgId:               job.Args.OrgId,
		Status:              enums.ServiceStatus(job.Args.Status),
		CheckedAt:           job.Args.CheckedAt,
		StatusCode:          job.Args.StatusCode,
		Error:               job.Args.Error,
		ConsecutiveFailures: job.Args.ConsecutiveFailures,
	})
	if serviceErr != nil {
		return serviceErr
	}
	if len(emails) == 0 {
		return nil
	}

	params := make([]river.InsertManyParams, 0, len(emails))
	for _, e := range emails {
		params = append(params, river.InsertManyParams{Args: args.SendEmailArgs{To: e.To, Subject: e.Subject, Body: e.Body}})
	}
	if _, err := river.ClientFromContext[pgx.Tx](ctx).InsertMany(ctx, params); err != nil {
		logger.Errorf("failed to queue %d alert emails for service %s : %v", len(params), job.Args.ServiceId, err)
		return err
	}
	logger.Infof("queued %d alert emails for service %s (%s)", len(params), job.Args.ServiceId, job.Args.Status)
	return nil
}

// sendEmailWorker sends one email; river retries it with backoff if the mail server fails
func sendEmailWorker(ctx custom.Context, job *river.Job[args.SendEmailArgs]) error {
	if err := email.Send(job.Args.To, job.Args.Subject, job.Args.Body); err != nil {
		logger.Errorf("failed to send email to %s (attempt %d of %d) : %v", job.Args.To, job.Attempt, job.MaxAttempts, err)
		return err
	}
	logger.Infof("sent email to %s: %s", job.Args.To, job.Args.Subject)
	return nil
}

func NewSendAlertWorker() river.Worker[args.SendAlertArgs] {
	return async_middleware.WithSystemWorker(sendAlertWorker)
}

func NewSendEmailWorker() river.Worker[args.SendEmailArgs] {
	return &emailWorker{Worker: async_middleware.WithSystemWorker(sendEmailWorker)}
}

// emailWorker gives each email attempt its own time limit, longer than the SMTP timeout
type emailWorker struct {
	river.Worker[args.SendEmailArgs]
}

func (w *emailWorker) Timeout(*river.Job[args.SendEmailArgs]) time.Duration {
	return time.Minute
}
