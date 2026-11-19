package daily_stats_repo

import (
	"sync"

	"pingwise/common/custom"
	"pingwise/internal/models"
	"pingwise/internal/repo"
)

type DailyStatsRepo struct {
	*repo.BaseRepo
}

var (
	dailyStatsRepo     *DailyStatsRepo
	dailyStatsRepoOnce sync.Once
)

func GetDailyStatsRepo() *DailyStatsRepo {
	dailyStatsRepoOnce.Do(func() {
		dailyStatsRepo = &DailyStatsRepo{
			BaseRepo: repo.GetBaseRepo(),
		}
	})
	return dailyStatsRepo
}

// ListByServiceIdSince returns the service's daily rows from sinceDay (UTC, inclusive), oldest first.
// Days without checks have no row.
func (r *DailyStatsRepo) ListByServiceIdSince(ctx custom.Context, serviceId, orgId, sinceDay string) ([]models.ServiceDailyStat, *repo.RepoErr) {
	var stats []models.ServiceDailyStat
	err := r.GetDb(ctx).Where("service_id = ? AND org_id = ? AND day >= ?", serviceId, orgId, sinceDay).
		Order("day").Find(&stats).Error
	if err != nil {
		return nil, repo.HandleErr(err)
	}
	return stats, nil
}

// ListByOrgIdSince returns the daily rows of every non-deleted service in the org from sinceDay
// (UTC, inclusive), oldest first
func (r *DailyStatsRepo) ListByOrgIdSince(ctx custom.Context, orgId, sinceDay string) ([]models.ServiceDailyStat, *repo.RepoErr) {
	var stats []models.ServiceDailyStat
	err := r.GetDb(ctx).
		Where("org_id = ? AND day >= ?", orgId, sinceDay).
		Where("service_id IN (?)", r.GetDb(ctx).Model(&models.Service{}).Select("id").Where("org_id = ?", orgId)).
		Order("service_id, day").Find(&stats).Error
	if err != nil {
		return nil, repo.HandleErr(err)
	}
	return stats, nil
}

// AddCheck adds one check to the service's row for the given UTC day, creating the row if needed
func (r *DailyStatsRepo) AddCheck(ctx custom.Context, serviceId, orgId, day string, isUp bool, responseTimeMs int) *repo.RepoErr {
	upChecks := 0
	if isUp {
		upChecks = 1
	}
	err := r.GetDb(ctx).Exec(`
		INSERT INTO service_daily_stats (service_id, org_id, day, total_checks, up_checks, total_response_ms, max_response_ms)
		VALUES (?, ?, ?, 1, ?, ?, ?)
		ON CONFLICT (service_id, day) DO UPDATE SET
			total_checks      = service_daily_stats.total_checks + 1,
			up_checks         = service_daily_stats.up_checks + EXCLUDED.up_checks,
			total_response_ms = service_daily_stats.total_response_ms + EXCLUDED.total_response_ms,
			max_response_ms   = GREATEST(service_daily_stats.max_response_ms, EXCLUDED.max_response_ms)`,
		serviceId, orgId, day, upChecks, responseTimeMs, responseTimeMs).Error
	if err != nil {
		return repo.HandleErr(err)
	}
	return nil
}
