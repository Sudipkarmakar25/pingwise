package notification_service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"pingwise/common/custom"
	"pingwise/common/email"
	"pingwise/internal/models/enums"
	"pingwise/internal/repo"
	"pingwise/internal/repo/alert_subscription_repo"
	"pingwise/internal/repo/service_repo"
	"pingwise/internal/service"
)

// Alert is a service going down or coming back up
type Alert struct {
	ServiceId           string
	OrgId               string
	Status              enums.ServiceStatus
	CheckedAt           time.Time
	StatusCode          *int
	Error               *string
	ConsecutiveFailures int
}

// Email is one message for one recipient
type Email struct {
	To      string
	Subject string
	Body    string
}

type Api interface {
	BuildAlertEmails(ctx custom.Context, alert *Alert) ([]Email, *service.Error)
}

type NotificationService struct {
	serviceRepo           *service_repo.ServiceRepo
	alertSubscriptionRepo *alert_subscription_repo.AlertSubscriptionRepo
}

var (
	notificationService     *NotificationService
	notificationServiceOnce sync.Once
)

func GetNotificationService() Api {
	notificationServiceOnce.Do(func() {
		notificationService = &NotificationService{
			serviceRepo:           service_repo.GetServiceRepo(),
			alertSubscriptionRepo: alert_subscription_repo.GetAlertSubscriptionRepo(),
		}
	})
	return notificationService
}

// BuildAlertEmails returns one email for every member who gets alerts for the service.
// It returns none when the service was deleted in the meantime.
func (s *NotificationService) BuildAlertEmails(ctx custom.Context, alert *Alert) ([]Email, *service.Error) {
	svc, repoErr := s.serviceRepo.GetByIdAndOrgId(ctx, alert.ServiceId, alert.OrgId)
	if repoErr != nil {
		if repoErr.Type == repo.NotFound {
			return nil, nil
		}
		return nil, service.HandleRepoErr(repoErr)
	}
	recipients, repoErr := s.alertSubscriptionRepo.ListRecipients(ctx, svc.Id, svc.OrgId)
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}

	when := alert.CheckedAt.UTC().Format("2 Jan 2006, 15:04 UTC")
	var subject string
	var body strings.Builder
	if alert.Status == enums.ServiceStatusDown {
		subject = fmt.Sprintf("[Pingwise] %s is down", svc.Name)
		fmt.Fprintf(&body, "%s is down.\n\n", svc.Name)
		fmt.Fprintf(&body, "URL:      %s %s\n", svc.Method, svc.Url)
		fmt.Fprintf(&body, "Reason:   %s\n", failureReason(alert.Error, alert.StatusCode, svc.ExpectedStatus, svc.TimeoutSeconds))
		fmt.Fprintf(&body, "Since:    %s\n", when)
		fmt.Fprintf(&body, "Failures: %d in a row\n", alert.ConsecutiveFailures)
	} else {
		subject = fmt.Sprintf("[Pingwise] %s has recovered", svc.Name)
		fmt.Fprintf(&body, "%s is up again.\n\n", svc.Name)
		fmt.Fprintf(&body, "URL:      %s %s\n", svc.Method, svc.Url)
		if alert.StatusCode != nil {
			fmt.Fprintf(&body, "Status:   %d\n", *alert.StatusCode)
		}
		fmt.Fprintf(&body, "Up since: %s\n", when)
	}
	if appUrl := email.AppUrl(); appUrl != "" {
		fmt.Fprintf(&body, "\nView service: %s/services/%s\n", appUrl, svc.Id)
	}
	body.WriteString("\n--\nYou get this email because alerts are turned on for you for this service in Pingwise.\n")

	emails := make([]Email, 0, len(recipients))
	for _, user := range recipients {
		emails = append(emails, Email{To: user.Email, Subject: subject, Body: body.String()})
	}
	return emails, nil
}

// failureReason turns the stored error code into a sentence
func failureReason(code *string, statusCode *int, expectedStatus, timeoutSeconds int) string {
	if code == nil {
		return "Check failed"
	}
	switch *code {
	case "status_mismatch":
		if statusCode != nil {
			return fmt.Sprintf("Unexpected status code (got %d, expected %d)", *statusCode, expectedStatus)
		}
		return "Unexpected status code"
	case "timeout":
		return fmt.Sprintf("Timed out (no response within %d s)", timeoutSeconds)
	case "dns_failure":
		return "DNS lookup failed"
	case "connection_refused":
		return "Connection refused"
	case "tls_error":
		return "TLS or certificate error"
	case "too_many_redirects":
		return "Too many redirects"
	case "token_decrypt_failed":
		return "Couldn't read the saved token"
	default:
		return "Request failed"
	}
}
