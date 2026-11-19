package service_repo

import (
	"database/sql"
	"sync"
	"time"

	"pingwise/common/custom"
	"pingwise/internal/models"
	"pingwise/internal/models/enums"
	"pingwise/internal/repo"
)

type ServiceRepo struct {
	*repo.BaseRepo
}

var (
	serviceRepo     *ServiceRepo
	serviceRepoOnce sync.Once
)

func GetServiceRepo() *ServiceRepo {
	serviceRepoOnce.Do(func() {
		serviceRepo = &ServiceRepo{
			BaseRepo: repo.GetBaseRepo(),
		}
	})
	return serviceRepo
}

func (r *ServiceRepo) Create(ctx custom.Context, svc *models.Service) *repo.RepoErr {
	if err := r.GetDb(ctx).Create(svc).Error; err != nil {
		return repo.HandleErr(err)
	}
	return nil
}

func (r *ServiceRepo) GetByIdAndOrgId(ctx custom.Context, id, orgId string) (*models.Service, *repo.RepoErr) {
	var svc models.Service
	if err := r.GetDb(ctx).Where("id = ? AND org_id = ?", id, orgId).First(&svc).Error; err != nil {
		return nil, repo.HandleErr(err)
	}
	return &svc, nil
}

func (r *ServiceRepo) ListByOrgId(ctx custom.Context, orgId string) ([]models.Service, *repo.RepoErr) {
	var services []models.Service
	if err := r.GetDb(ctx).Where("org_id = ?", orgId).Order("created_at").Find(&services).Error; err != nil {
		return nil, repo.HandleErr(err)
	}
	return services, nil
}

func (r *ServiceRepo) Update(ctx custom.Context, id, orgId string, updates map[string]interface{}) *repo.RepoErr {
	err := r.GetDb(ctx).Model(&models.Service{}).Where("id = ? AND org_id = ?", id, orgId).Updates(updates).Error
	if err != nil {
		return repo.HandleErr(err)
	}
	return nil
}

// ClaimDueServices marks every active service whose interval has passed as scheduled for the
// current minute and returns them. Times are truncated to the minute so that a tick running a
// few milliseconds early or late does not make a service skip an interval.
func (r *ServiceRepo) ClaimDueServices(ctx custom.Context) ([]models.Service, *repo.RepoErr) {
	var services []models.Service
	err := r.GetDb(ctx).Raw(`
		UPDATE services SET last_scheduled_at = date_trunc('minute', now())
		WHERE is_active AND deleted_at IS NULL
		AND (last_scheduled_at IS NULL
			OR last_scheduled_at + make_interval(mins => interval_minutes) <= date_trunc('minute', now()))
		RETURNING id, org_id`).Scan(&services).Error
	if err != nil {
		return nil, repo.HandleErr(err)
	}
	return services, nil
}

// StatusUpdate is a service's status before and after a check was recorded
type StatusUpdate struct {
	OldStatus           enums.ServiceStatus
	NewStatus           enums.ServiceStatus
	ConsecutiveFailures int
}

// RecordCheck stores the outcome of a check on the service. A pass resets the failure count and
// marks it up; a failure marks it down only once failure_threshold failures happen in a row.
// It returns the status before and after, read and written in the same statement.
func (r *ServiceRepo) RecordCheck(ctx custom.Context, id string, isUp bool, checkedAt time.Time, responseTimeMs int) (*StatusUpdate, *repo.RepoErr) {
	var update StatusUpdate
	err := r.GetDb(ctx).Raw(`
		UPDATE services s SET
			current_status = CASE
				WHEN @is_up THEN 'up'
				WHEN s.consecutive_failures + 1 >= s.failure_threshold THEN 'down'
				ELSE s.current_status END,
			consecutive_failures = CASE WHEN @is_up THEN 0 ELSE s.consecutive_failures + 1 END,
			last_checked_at = @checked_at,
			last_response_time_ms = @response_time_ms
		FROM (SELECT id, current_status FROM services WHERE id = @id FOR UPDATE) old
		WHERE s.id = old.id
		RETURNING old.current_status AS old_status, s.current_status AS new_status, s.consecutive_failures`,
		sql.Named("is_up", isUp), sql.Named("checked_at", checkedAt),
		sql.Named("response_time_ms", responseTimeMs), sql.Named("id", id)).Scan(&update).Error
	if err != nil {
		return nil, repo.HandleErr(err)
	}
	return &update, nil
}

// SoftDelete sets deleted_at; the scheduler and every other query then skip the service
func (r *ServiceRepo) SoftDelete(ctx custom.Context, id, orgId string) *repo.RepoErr {
	if err := r.GetDb(ctx).Where("id = ? AND org_id = ?", id, orgId).Delete(&models.Service{}).Error; err != nil {
		return repo.HandleErr(err)
	}
	return nil
}
