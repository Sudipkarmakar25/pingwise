package alert_subscription_repo

import (
	"sync"

	"pingwise/common/custom"
	"pingwise/internal/models"
	"pingwise/internal/repo"
)

type AlertSubscriptionRepo struct {
	*repo.BaseRepo
}

var (
	alertSubscriptionRepo     *AlertSubscriptionRepo
	alertSubscriptionRepoOnce sync.Once
)

func GetAlertSubscriptionRepo() *AlertSubscriptionRepo {
	alertSubscriptionRepoOnce.Do(func() {
		alertSubscriptionRepo = &AlertSubscriptionRepo{
			BaseRepo: repo.GetBaseRepo(),
		}
	})
	return alertSubscriptionRepo
}

// ListByUserId returns every row of the user, on and off
func (r *AlertSubscriptionRepo) ListByUserId(ctx custom.Context, userId, orgId string) ([]models.AlertSubscription, *repo.RepoErr) {
	var subscriptions []models.AlertSubscription
	err := r.GetDb(ctx).Where("user_id = ? AND org_id = ?", userId, orgId).Find(&subscriptions).Error
	if err != nil {
		return nil, repo.HandleErr(err)
	}
	return subscriptions, nil
}

// ListByServiceId returns every row of the service, on and off
func (r *AlertSubscriptionRepo) ListByServiceId(ctx custom.Context, serviceId, orgId string) ([]models.AlertSubscription, *repo.RepoErr) {
	var subscriptions []models.AlertSubscription
	err := r.GetDb(ctx).Where("service_id = ? AND org_id = ?", serviceId, orgId).Find(&subscriptions).Error
	if err != nil {
		return nil, repo.HandleErr(err)
	}
	return subscriptions, nil
}

// ListRecipients returns the members who get alert emails for the service: those with the
// setting turned on, plus owners who have not turned it off
func (r *AlertSubscriptionRepo) ListRecipients(ctx custom.Context, serviceId, orgId string) ([]models.User, *repo.RepoErr) {
	var users []models.User
	err := r.GetDb(ctx).Raw(`
		SELECT u.* FROM users u
		LEFT JOIN alert_subscriptions a ON a.user_id = u.id AND a.service_id = ?
		WHERE u.org_id = ? AND u.deleted_at IS NULL
		AND COALESCE(a.is_enabled, u.role = 'owner')
		ORDER BY u.created_at`, serviceId, orgId).Scan(&users).Error
	if err != nil {
		return nil, repo.HandleErr(err)
	}
	return users, nil
}

// Upsert creates the row the first time and only updates is_enabled and updated_by after that
func (r *AlertSubscriptionRepo) Upsert(ctx custom.Context, subscription *models.AlertSubscription) *repo.RepoErr {
	err := r.GetDb(ctx).Raw(`
		INSERT INTO alert_subscriptions (org_id, user_id, service_id, is_enabled, updated_by)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id, service_id) DO UPDATE SET
			is_enabled = EXCLUDED.is_enabled,
			updated_by = EXCLUDED.updated_by,
			updated_at = now()
		RETURNING *`,
		subscription.OrgId, subscription.UserId, subscription.ServiceId, subscription.IsEnabled, subscription.UpdatedBy).
		Scan(subscription).Error
	if err != nil {
		return repo.HandleErr(err)
	}
	return nil
}
