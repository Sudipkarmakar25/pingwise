package monitor_service

import (
	"math"
	"time"

	"pingwise/internal/models"
)

const maskPrefix = "••••"

type ServiceResponse struct {
	*models.Service
	AuthTokenMasked *string `json:"auth_token_masked"`
	Notify          bool    `json:"notify"` // whether the caller gets alert emails for this service
}

func newServiceResponse(svc *models.Service, notify bool) *ServiceResponse {
	res := &ServiceResponse{Service: svc, Notify: notify}
	if svc.AuthTokenEncrypted != nil {
		masked := maskPrefix
		if svc.AuthTokenLast4 != nil {
			masked += *svc.AuthTokenLast4
		}
		res.AuthTokenMasked = &masked
	}
	return res
}

type DailyStatResponse struct {
	Day           string `json:"day"` // UTC date, YYYY-MM-DD
	TotalChecks   int    `json:"total_checks"`
	UpChecks      int    `json:"up_checks"`
	AvgResponseMs int    `json:"avg_response_ms"`
	MaxResponseMs int    `json:"max_response_ms"`
}

// ServiceDetailResponse holds the service and the data for its uptime bars. daily_stats has a
// row only for days with checks; days in between should be shown as having no data.
type ServiceDetailResponse struct {
	Service          *ServiceResponse    `json:"service"`
	FromDay          string              `json:"from_day"`
	ToDay            string              `json:"to_day"`
	DailyStats       []DailyStatResponse `json:"daily_stats"`
	UptimePercentage *float64            `json:"uptime_percentage"` // nil until the first check
}

func newServiceDetailResponse(svc *models.Service, notify bool, stats []models.ServiceDailyStat, fromDay, toDay string) *ServiceDetailResponse {
	summary := summariseStats(stats)
	return &ServiceDetailResponse{
		Service:          newServiceResponse(svc, notify),
		FromDay:          fromDay,
		ToDay:            toDay,
		DailyStats:       summary.DailyStats,
		UptimePercentage: summary.UptimePercentage,
	}
}

type StatsSummary struct {
	DailyStats       []DailyStatResponse `json:"daily_stats"`
	UptimePercentage *float64            `json:"uptime_percentage"` // nil until the first check
}

// OrgStatsResponse holds the daily stats of every service in the org, keyed by service id.
// A service without checks in the period has no entry.
type OrgStatsResponse struct {
	FromDay  string                   `json:"from_day"`
	ToDay    string                   `json:"to_day"`
	Services map[string]*StatsSummary `json:"services"`
}

func newOrgStatsResponse(stats []models.ServiceDailyStat, fromDay, toDay string) *OrgStatsResponse {
	byService := map[string][]models.ServiceDailyStat{}
	for _, stat := range stats {
		byService[stat.ServiceId] = append(byService[stat.ServiceId], stat)
	}
	res := &OrgStatsResponse{FromDay: fromDay, ToDay: toDay, Services: make(map[string]*StatsSummary, len(byService))}
	for serviceId, serviceStats := range byService {
		res.Services[serviceId] = summariseStats(serviceStats)
	}
	return res
}

// summariseStats converts daily rows to responses and adds the uptime over all of them
func summariseStats(stats []models.ServiceDailyStat) *StatsSummary {
	summary := &StatsSummary{DailyStats: make([]DailyStatResponse, 0, len(stats))}
	totalChecks, upChecks := 0, 0
	for _, stat := range stats {
		summary.DailyStats = append(summary.DailyStats, DailyStatResponse{
			Day:           stat.Day.Format(time.DateOnly),
			TotalChecks:   stat.TotalChecks,
			UpChecks:      stat.UpChecks,
			AvgResponseMs: int(stat.TotalResponseMs / int64(stat.TotalChecks)),
			MaxResponseMs: stat.MaxResponseMs,
		})
		totalChecks += stat.TotalChecks
		upChecks += stat.UpChecks
	}
	if totalChecks > 0 {
		uptime := math.Round(float64(upChecks)/float64(totalChecks)*10000) / 100
		summary.UptimePercentage = &uptime
	}
	return summary
}

type TimelineBucket struct {
	Start         string `json:"start"` // RFC 3339, UTC
	TotalChecks   int    `json:"total_checks"`
	UpChecks      int    `json:"up_checks"`
	AvgResponseMs int    `json:"avg_response_ms"`
	MaxResponseMs int    `json:"max_response_ms"`
}

// TimelineResponse holds the slots that had checks. From and To are the starts of the first and
// last slot of the window, so the caller can draw empty slots as having no data.
type TimelineResponse struct {
	Granularity   string           `json:"granularity"`
	BucketMinutes int              `json:"bucket_minutes"`
	From          string           `json:"from"`
	To            string           `json:"to"`
	Buckets       []TimelineBucket `json:"buckets"`
}

type ListChecksResponse struct {
	Checks     []models.CheckResult `json:"checks"`
	NextCursor string               `json:"next_cursor"` // empty when there are no more pages
}
