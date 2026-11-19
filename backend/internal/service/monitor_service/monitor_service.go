package monitor_service

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"pingwise/common/custom"
	"pingwise/common/encryption"
	"pingwise/common/logger"
	"pingwise/internal/models"
	"pingwise/internal/models/enums"
	"pingwise/internal/repo/alert_subscription_repo"
	"pingwise/internal/repo/check_result_repo"
	"pingwise/internal/repo/daily_stats_repo"
	"pingwise/internal/repo/service_repo"
	"pingwise/internal/service"
)

const (
	defaultMethod           = http.MethodGet
	defaultExpectedStatus   = http.StatusOK
	defaultIntervalMinutes  = 1
	defaultTimeoutSeconds   = 10
	defaultFailureThreshold = 2

	// tokens shorter than this are masked fully, so the last 4 characters do not give most of it away
	minTokenLengthToShowLast4 = 12

	statsDays             = 90 // days of uptime bars on the service detail
	defaultChecksPageSize = 50
)

var headerNamePattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

type Api interface {
	CreateService(ctx custom.Context, req *CreateServiceRequest) (*ServiceResponse, *service.Error)
	ListServices(ctx custom.Context) ([]*ServiceResponse, *service.Error)
	GetService(ctx custom.Context, req *ServiceIdRequest) (*ServiceDetailResponse, *service.Error)
	ListChecks(ctx custom.Context, req *ListChecksRequest) (*ListChecksResponse, *service.Error)
	ListStats(ctx custom.Context) (*OrgStatsResponse, *service.Error)
	GetTimeline(ctx custom.Context, req *TimelineRequest) (*TimelineResponse, *service.Error)
	UpdateService(ctx custom.Context, req *UpdateServiceRequest) (*ServiceResponse, *service.Error)
	DeleteService(ctx custom.Context, req *ServiceIdRequest) *service.Error
}

type MonitorService struct {
	serviceRepo     *service_repo.ServiceRepo
	dailyStatsRepo  *daily_stats_repo.DailyStatsRepo
	checkResultRepo *check_result_repo.CheckResultRepo
	alertRepo       *alert_subscription_repo.AlertSubscriptionRepo
}

var (
	monitorService     *MonitorService
	monitorServiceOnce sync.Once
)

func GetMonitorService() Api {
	monitorServiceOnce.Do(func() {
		monitorService = &MonitorService{
			serviceRepo:     service_repo.GetServiceRepo(),
			dailyStatsRepo:  daily_stats_repo.GetDailyStatsRepo(),
			checkResultRepo: check_result_repo.GetCheckResultRepo(),
			alertRepo:       alert_subscription_repo.GetAlertSubscriptionRepo(),
		}
	})
	return monitorService
}

func (s *MonitorService) CreateService(ctx custom.Context, req *CreateServiceRequest) (*ServiceResponse, *service.Error) {
	method := valueOrDefault(req.Method, defaultMethod)
	if serviceErr := validateRequestBody(method, req.RequestBody); serviceErr != nil {
		return nil, serviceErr
	}
	if serviceErr := validateHeaders(req.AuthHeaderName, req.CustomHeaders); serviceErr != nil {
		return nil, serviceErr
	}
	if (req.AuthHeaderName == "") != (req.AuthToken == "") {
		return nil, service.NewError("auth_header_name and auth_token must be set together", http.StatusBadRequest)
	}

	svc := &models.Service{
		OrgId:            ctx.GetOrgId(),
		Name:             strings.TrimSpace(req.Name),
		Url:              req.Url,
		Method:           method,
		CustomHeaders:    req.CustomHeaders,
		ExpectedStatus:   valueOrDefault(req.ExpectedStatus, defaultExpectedStatus),
		IntervalMinutes:  valueOrDefault(req.IntervalMinutes, defaultIntervalMinutes),
		TimeoutSeconds:   valueOrDefault(req.TimeoutSeconds, defaultTimeoutSeconds),
		FailureThreshold: valueOrDefault(req.FailureThreshold, defaultFailureThreshold),
		IsActive:         true,
		CurrentStatus:    enums.ServiceStatusUnknown,
		CreatedBy:        ctx.GetUserId(),
	}
	if svc.CustomHeaders == nil {
		svc.CustomHeaders = map[string]string{}
	}
	if req.RequestBody != "" {
		svc.RequestBody = &req.RequestBody
	}
	if req.AuthToken != "" {
		encrypted, last4, keyVersion, serviceErr := encryptAuthToken(req.AuthToken)
		if serviceErr != nil {
			return nil, serviceErr
		}
		svc.AuthHeaderName = &req.AuthHeaderName
		svc.AuthTokenEncrypted = encrypted
		svc.AuthTokenLast4 = last4
		svc.AuthKeyVersion = &keyVersion
	}

	if repoErr := s.serviceRepo.Create(ctx, svc); repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	// owners get alerts by default; everyone else starts with them off
	return newServiceResponse(svc, ctx.GetUser().Role == enums.UserRoleOwner), nil
}

func (s *MonitorService) ListServices(ctx custom.Context) ([]*ServiceResponse, *service.Error) {
	services, repoErr := s.serviceRepo.ListByOrgId(ctx, ctx.GetOrgId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	isNotified, serviceErr := s.notifiedServices(ctx)
	if serviceErr != nil {
		return nil, serviceErr
	}
	res := make([]*ServiceResponse, 0, len(services))
	for i := range services {
		res = append(res, newServiceResponse(&services[i], isNotified(services[i].Id)))
	}
	return res, nil
}

// GetService returns the service with its daily stats for the last 90 UTC days, today included
func (s *MonitorService) GetService(ctx custom.Context, req *ServiceIdRequest) (*ServiceDetailResponse, *service.Error) {
	svc, repoErr := s.serviceRepo.GetByIdAndOrgId(ctx, req.Id, ctx.GetOrgId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	today := time.Now().UTC()
	fromDay := today.AddDate(0, 0, -(statsDays - 1)).Format(time.DateOnly)
	stats, repoErr := s.dailyStatsRepo.ListByServiceIdSince(ctx, svc.Id, svc.OrgId, fromDay)
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	isNotified, serviceErr := s.notifiedServices(ctx)
	if serviceErr != nil {
		return nil, serviceErr
	}
	return newServiceDetailResponse(svc, isNotified(svc.Id), stats, fromDay, today.Format(time.DateOnly)), nil
}

// timelineWindows sets, for each granularity, the slot size in minutes and how many slots are returned
var timelineWindows = map[string]struct{ bucketMinutes, buckets int }{
	"minute":       {1, 180},             // last 3 hours
	"five_minutes": {5, 288},             // last 24 hours
	"day":          {24 * 60, statsDays}, // last 90 days
}

// GetTimeline returns the service's checks grouped by minute, by 5 minutes or by UTC day. Minute
// and 5-minute slots come from individual checks; days come from the daily stats.
func (s *MonitorService) GetTimeline(ctx custom.Context, req *TimelineRequest) (*TimelineResponse, *service.Error) {
	svc, repoErr := s.serviceRepo.GetByIdAndOrgId(ctx, req.Id, ctx.GetOrgId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	window := timelineWindows[req.Granularity]
	step := time.Duration(window.bucketMinutes) * time.Minute
	to := time.Now().UTC().Truncate(step) // UTC midnight for days
	from := to.Add(-step * time.Duration(window.buckets-1))

	res := &TimelineResponse{
		Granularity:   req.Granularity,
		BucketMinutes: window.bucketMinutes,
		From:          from.Format(time.RFC3339),
		To:            to.Format(time.RFC3339),
		Buckets:       []TimelineBucket{},
	}

	if req.Granularity == "day" {
		stats, repoErr := s.dailyStatsRepo.ListByServiceIdSince(ctx, svc.Id, svc.OrgId, from.Format(time.DateOnly))
		if repoErr != nil {
			return nil, service.HandleRepoErr(repoErr)
		}
		for _, stat := range stats {
			res.Buckets = append(res.Buckets, TimelineBucket{
				Start:         stat.Day.UTC().Format(time.RFC3339),
				TotalChecks:   stat.TotalChecks,
				UpChecks:      stat.UpChecks,
				AvgResponseMs: int(stat.TotalResponseMs / int64(stat.TotalChecks)),
				MaxResponseMs: stat.MaxResponseMs,
			})
		}
		return res, nil
	}

	buckets, repoErr := s.checkResultRepo.ListBuckets(ctx, svc.Id, svc.OrgId, window.bucketMinutes, from)
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	for _, b := range buckets {
		res.Buckets = append(res.Buckets, TimelineBucket{
			Start:         b.BucketStart.UTC().Format(time.RFC3339),
			TotalChecks:   b.TotalChecks,
			UpChecks:      b.UpChecks,
			AvgResponseMs: b.AvgResponseMs,
			MaxResponseMs: b.MaxResponseMs,
		})
	}
	return res, nil
}

// ListStats returns the last 90 UTC days of daily stats for every service in the org, for the
// uptime bars on the dashboard
func (s *MonitorService) ListStats(ctx custom.Context) (*OrgStatsResponse, *service.Error) {
	today := time.Now().UTC()
	fromDay := today.AddDate(0, 0, -(statsDays - 1)).Format(time.DateOnly)
	stats, repoErr := s.dailyStatsRepo.ListByOrgIdSince(ctx, ctx.GetOrgId(), fromDay)
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	return newOrgStatsResponse(stats, fromDay, today.Format(time.DateOnly)), nil
}

func (s *MonitorService) ListChecks(ctx custom.Context, req *ListChecksRequest) (*ListChecksResponse, *service.Error) {
	var before *time.Time
	if req.Cursor != "" {
		cursor, err := time.Parse(time.RFC3339Nano, req.Cursor)
		if err != nil {
			return nil, service.NewError("invalid cursor", http.StatusBadRequest)
		}
		before = &cursor
	}
	svc, repoErr := s.serviceRepo.GetByIdAndOrgId(ctx, req.Id, ctx.GetOrgId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}

	limit := valueOrDefault(req.Limit, defaultChecksPageSize)
	checks, repoErr := s.checkResultRepo.ListByServiceId(ctx, svc.Id, svc.OrgId, before, limit)
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	res := &ListChecksResponse{Checks: checks}
	if len(checks) == limit {
		res.NextCursor = checks[len(checks)-1].CheckedAt.UTC().Format(time.RFC3339Nano)
	}
	return res, nil
}

func (s *MonitorService) UpdateService(ctx custom.Context, req *UpdateServiceRequest) (*ServiceResponse, *service.Error) {
	svc, repoErr := s.serviceRepo.GetByIdAndOrgId(ctx, req.Id, ctx.GetOrgId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}

	// any change starts the failure count again, so old failures do not count against the new settings
	updates := map[string]interface{}{"consecutive_failures": 0}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Url != nil {
		updates["url"] = *req.Url
	}
	if req.ExpectedStatus != nil {
		updates["expected_status"] = *req.ExpectedStatus
	}
	if req.IntervalMinutes != nil {
		updates["interval_minutes"] = *req.IntervalMinutes
	}
	if req.TimeoutSeconds != nil {
		updates["timeout_seconds"] = *req.TimeoutSeconds
	}
	if req.FailureThreshold != nil {
		updates["failure_threshold"] = *req.FailureThreshold
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// method and body are validated together using the values the service will end up with
	method := svc.Method
	if req.Method != nil {
		method = *req.Method
		updates["method"] = method
	}
	body := ""
	if svc.RequestBody != nil {
		body = *svc.RequestBody
	}
	if req.RequestBody != nil {
		body = *req.RequestBody
		if body == "" {
			updates["request_body"] = nil
		} else {
			updates["request_body"] = body
		}
	}
	if serviceErr := validateRequestBody(method, body); serviceErr != nil {
		return nil, serviceErr
	}

	authHeaderName := ""
	if svc.AuthHeaderName != nil {
		authHeaderName = *svc.AuthHeaderName
	}
	if req.AuthHeaderName != nil {
		authHeaderName = *req.AuthHeaderName
	}
	customHeaders := svc.CustomHeaders
	if req.CustomHeaders != nil {
		customHeaders = req.CustomHeaders
		customHeadersJson, err := json.Marshal(customHeaders)
		if err != nil {
			return nil, service.NewError(service.InvalidPayload, http.StatusBadRequest)
		}
		updates["custom_headers"] = string(customHeadersJson)
	}

	switch {
	case req.AuthToken != nil && *req.AuthToken == "":
		// remove auth
		authHeaderName = ""
		updates["auth_header_name"] = nil
		updates["auth_token_encrypted"] = nil
		updates["auth_token_last4"] = nil
		updates["auth_key_version"] = nil
	case req.AuthToken != nil:
		if authHeaderName == "" {
			return nil, service.NewError("auth_header_name is required with auth_token", http.StatusBadRequest)
		}
		encrypted, last4, keyVersion, serviceErr := encryptAuthToken(*req.AuthToken)
		if serviceErr != nil {
			return nil, serviceErr
		}
		updates["auth_header_name"] = authHeaderName
		updates["auth_token_encrypted"] = encrypted
		updates["auth_token_last4"] = last4
		updates["auth_key_version"] = keyVersion
	case req.AuthHeaderName != nil:
		// rename the header of the existing token
		if svc.AuthTokenEncrypted == nil {
			return nil, service.NewError("auth_token is required with auth_header_name", http.StatusBadRequest)
		}
		if authHeaderName == "" {
			return nil, service.NewError("send auth_token as \"\" to remove auth", http.StatusBadRequest)
		}
		updates["auth_header_name"] = authHeaderName
	}
	if serviceErr := validateHeaders(authHeaderName, customHeaders); serviceErr != nil {
		return nil, serviceErr
	}

	if repoErr := s.serviceRepo.Update(ctx, svc.Id, svc.OrgId, updates); repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	svc, repoErr = s.serviceRepo.GetByIdAndOrgId(ctx, svc.Id, svc.OrgId)
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	isNotified, serviceErr := s.notifiedServices(ctx)
	if serviceErr != nil {
		return nil, serviceErr
	}
	return newServiceResponse(svc, isNotified(svc.Id)), nil
}

func (s *MonitorService) DeleteService(ctx custom.Context, req *ServiceIdRequest) *service.Error {
	if _, repoErr := s.serviceRepo.GetByIdAndOrgId(ctx, req.Id, ctx.GetOrgId()); repoErr != nil {
		return service.HandleRepoErr(repoErr)
	}
	if repoErr := s.serviceRepo.SoftDelete(ctx, req.Id, ctx.GetOrgId()); repoErr != nil {
		return service.HandleRepoErr(repoErr)
	}
	return nil
}

// notifiedServices tells whether the caller gets alert emails for a service. A row decides it;
// without one, owners get them and editors and viewers do not.
func (s *MonitorService) notifiedServices(ctx custom.Context) (func(serviceId string) bool, *service.Error) {
	subscriptions, repoErr := s.alertRepo.ListByUserId(ctx, ctx.GetUserId(), ctx.GetOrgId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	byServiceId := make(map[string]bool, len(subscriptions))
	for _, subscription := range subscriptions {
		byServiceId[subscription.ServiceId] = subscription.IsEnabled
	}
	isOwner := ctx.GetUser().Role == enums.UserRoleOwner
	return func(serviceId string) bool {
		if enabled, ok := byServiceId[serviceId]; ok {
			return enabled
		}
		return isOwner
	}, nil
}

// encryptAuthToken returns the encrypted token, its last 4 characters for display, and the key version
func encryptAuthToken(token string) ([]byte, *string, int, *service.Error) {
	encrypted, keyVersion, err := encryption.Encrypt(token)
	if err != nil {
		logger.Errorf("failed to encrypt auth token : %v", err)
		return nil, nil, 0, service.NewError(service.SomethingUnExpectedErr, http.StatusInternalServerError)
	}
	var last4 *string
	if len(token) >= minTokenLengthToShowLast4 {
		l := token[len(token)-4:]
		last4 = &l
	}
	return encrypted, last4, keyVersion, nil
}

func validateRequestBody(method, body string) *service.Error {
	if body != "" && method != http.MethodPost {
		return service.NewError("request_body is allowed only with POST", http.StatusBadRequest)
	}
	return nil
}

func validateHeaders(authHeaderName string, customHeaders map[string]string) *service.Error {
	if authHeaderName != "" && !headerNamePattern.MatchString(authHeaderName) {
		return service.NewError("auth_header_name is not a valid header name", http.StatusBadRequest)
	}
	for name := range customHeaders {
		if !headerNamePattern.MatchString(name) {
			return service.NewError("custom header \""+name+"\" is not a valid header name", http.StatusBadRequest)
		}
		if authHeaderName != "" && strings.EqualFold(name, authHeaderName) {
			return service.NewError("custom header \""+name+"\" duplicates the auth header", http.StatusBadRequest)
		}
	}
	return nil
}

func valueOrDefault[T comparable](value, defaultValue T) T {
	var zero T
	if value == zero {
		return defaultValue
	}
	return value
}
