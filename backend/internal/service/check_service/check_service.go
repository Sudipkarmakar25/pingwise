package check_service

import (
	"sync"
	"time"

	"pingwise/common/custom"
	"pingwise/common/database"
	"pingwise/common/encryption"
	"pingwise/common/logger"
	"pingwise/internal/models"
	"pingwise/internal/models/enums"
	"pingwise/internal/repo"
	"pingwise/internal/repo/check_result_repo"
	"pingwise/internal/repo/daily_stats_repo"
	"pingwise/internal/repo/service_repo"
	"pingwise/internal/service"
)

type Api interface {
	ClaimDueServices(ctx custom.Context) ([]models.Service, *service.Error)
	RunCheck(ctx custom.Context, serviceId, orgId string) (*StatusChange, *service.Error)
}

// StatusChange is returned by RunCheck when a service went down or came back up, so an alert
// can be sent. A new service's first passing check is not a change.
type StatusChange struct {
	ServiceId           string
	OrgId               string
	Status              enums.ServiceStatus // up or down
	CheckedAt           time.Time
	StatusCode          *int
	Error               *string
	ConsecutiveFailures int
}

type CheckService struct {
	serviceRepo     *service_repo.ServiceRepo
	checkResultRepo *check_result_repo.CheckResultRepo
	dailyStatsRepo  *daily_stats_repo.DailyStatsRepo
}

var (
	checkService     *CheckService
	checkServiceOnce sync.Once
)

func GetCheckService() Api {
	checkServiceOnce.Do(func() {
		checkService = &CheckService{
			serviceRepo:     service_repo.GetServiceRepo(),
			checkResultRepo: check_result_repo.GetCheckResultRepo(),
			dailyStatsRepo:  daily_stats_repo.GetDailyStatsRepo(),
		}
	})
	return checkService
}

func (s *CheckService) ClaimDueServices(ctx custom.Context) ([]models.Service, *service.Error) {
	services, repoErr := s.serviceRepo.ClaimDueServices(ctx)
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	return services, nil
}

// RunCheck checks one service and saves the result, the daily stats and the new status together.
// It returns a StatusChange when the service went down or came back up, else nil.
func (s *CheckService) RunCheck(ctx custom.Context, serviceId, orgId string) (*StatusChange, *service.Error) {
	svc, repoErr := s.serviceRepo.GetByIdAndOrgId(ctx, serviceId, orgId)
	if repoErr != nil {
		if repoErr.Type == repo.NotFound {
			return nil, nil // deleted after it was scheduled
		}
		return nil, service.HandleRepoErr(repoErr)
	}
	if !svc.IsActive {
		return nil, nil // paused after it was scheduled
	}

	checkedAt := time.Now().UTC()
	var outcome *checkOutcome
	authToken, err := decryptAuthToken(svc)
	if err != nil {
		logger.Errorf("failed to decrypt auth token for service %s : %v", svc.Id, err)
		outcome = failed(errTokenDecryptFailed, 0)
	} else {
		outcome = performCheck(ctx, svc, authToken)
	}

	var update *service_repo.StatusUpdate
	txErr := database.WithTransaction(ctx, func() error {
		result := &models.CheckResult{
			ServiceId:      svc.Id,
			OrgId:          svc.OrgId,
			CheckedAt:      checkedAt,
			IsUp:           outcome.IsUp,
			StatusCode:     outcome.StatusCode,
			ResponseTimeMs: outcome.ResponseTimeMs,
			Error:          outcome.Error,
		}
		if repoErr := s.checkResultRepo.Create(ctx, result); repoErr != nil {
			return repoErr
		}
		if repoErr := s.dailyStatsRepo.AddCheck(ctx, svc.Id, svc.OrgId, checkedAt.Format(time.DateOnly),
			outcome.IsUp, outcome.ResponseTimeMs); repoErr != nil {
			return repoErr
		}
		var repoErr *repo.RepoErr
		update, repoErr = s.serviceRepo.RecordCheck(ctx, svc.Id, outcome.IsUp, checkedAt, outcome.ResponseTimeMs)
		if repoErr != nil {
			return repoErr
		}
		return nil
	})
	if txErr != nil {
		return nil, service.HandleErr(txErr)
	}

	errValue := ""
	if outcome.Error != nil {
		errValue = *outcome.Error
	}
	logger.Infof("checked service %s (%s): up=%v status=%v time=%dms %s",
		svc.Id, svc.Name, outcome.IsUp, valueOf(outcome.StatusCode), outcome.ResponseTimeMs, errValue)

	if !isAlertable(update) {
		return nil, nil
	}
	return &StatusChange{
		ServiceId:           svc.Id,
		OrgId:               svc.OrgId,
		Status:              update.NewStatus,
		CheckedAt:           checkedAt,
		StatusCode:          outcome.StatusCode,
		Error:               outcome.Error,
		ConsecutiveFailures: update.ConsecutiveFailures,
	}, nil
}

// isAlertable is true when the service went down, or came back up after being down
func isAlertable(update *service_repo.StatusUpdate) bool {
	if update == nil || update.OldStatus == update.NewStatus {
		return false
	}
	return update.NewStatus == enums.ServiceStatusDown ||
		(update.NewStatus == enums.ServiceStatusUp && update.OldStatus == enums.ServiceStatusDown)
}

// decryptAuthToken returns the plain token, or "" when the service has no auth
func decryptAuthToken(svc *models.Service) (string, error) {
	if svc.AuthTokenEncrypted == nil || svc.AuthKeyVersion == nil {
		return "", nil
	}
	return encryption.Decrypt(svc.AuthTokenEncrypted, *svc.AuthKeyVersion)
}

func valueOf(v *int) interface{} {
	if v == nil {
		return "-"
	}
	return *v
}
