package check_result_repo

import (
	"sync"
	"time"

	"pingwise/common/custom"
	"pingwise/internal/models"
	"pingwise/internal/repo"
)

type CheckResultRepo struct {
	*repo.BaseRepo
}

var (
	checkResultRepo     *CheckResultRepo
	checkResultRepoOnce sync.Once
)

func GetCheckResultRepo() *CheckResultRepo {
	checkResultRepoOnce.Do(func() {
		checkResultRepo = &CheckResultRepo{
			BaseRepo: repo.GetBaseRepo(),
		}
	})
	return checkResultRepo
}

// ListByServiceId returns the newest checks first. Pass the checked_at of the last row seen as
// before to get the next page.
func (r *CheckResultRepo) ListByServiceId(ctx custom.Context, serviceId, orgId string, before *time.Time,
	limit int) ([]models.CheckResult, *repo.RepoErr) {
	query := r.GetDb(ctx).Where("service_id = ? AND org_id = ?", serviceId, orgId)
	if before != nil {
		query = query.Where("checked_at < ?", *before)
	}
	var results []models.CheckResult
	if err := query.Order("checked_at DESC").Limit(limit).Find(&results).Error; err != nil {
		return nil, repo.HandleErr(err)
	}
	return results, nil
}

// CheckBucket is the checks of one service grouped into a fixed time slot
type CheckBucket struct {
	BucketStart   time.Time
	TotalChecks   int
	UpChecks      int
	AvgResponseMs int
	MaxResponseMs int
}

// ListBuckets groups the service's checks since `since` into slots of bucketMinutes, oldest first.
// Slots without checks are not returned.
func (r *CheckResultRepo) ListBuckets(ctx custom.Context, serviceId, orgId string, bucketMinutes int,
	since time.Time) ([]CheckBucket, *repo.RepoErr) {
	var buckets []CheckBucket
	err := r.GetDb(ctx).Raw(`
		SELECT date_bin(make_interval(mins => ?), checked_at, TIMESTAMPTZ '2000-01-01 00:00:00+00') AS bucket_start,
			count(*) AS total_checks,
			count(*) FILTER (WHERE is_up) AS up_checks,
			round(avg(response_time_ms))::int AS avg_response_ms,
			max(response_time_ms) AS max_response_ms
		FROM check_results
		WHERE service_id = ? AND org_id = ? AND checked_at >= ?
		GROUP BY 1
		ORDER BY 1`, bucketMinutes, serviceId, orgId, since).Scan(&buckets).Error
	if err != nil {
		return nil, repo.HandleErr(err)
	}
	return buckets, nil
}

func (r *CheckResultRepo) Create(ctx custom.Context, result *models.CheckResult) *repo.RepoErr {
	if err := r.GetDb(ctx).Create(result).Error; err != nil {
		return repo.HandleErr(err)
	}
	return nil
}
