package alert_service

import (
	"net/http"
	"sync"

	"pingwise/common/custom"
	"pingwise/internal/models"
	"pingwise/internal/models/enums"
	"pingwise/internal/repo"
	"pingwise/internal/repo/alert_subscription_repo"
	"pingwise/internal/repo/service_repo"
	"pingwise/internal/repo/user_repo"
	"pingwise/internal/service"
)

// roleRank orders roles so a higher role can change a lower role's alerts
var roleRank = map[enums.UserRole]int{
	enums.UserRoleOwner:  3,
	enums.UserRoleEditor: 2,
	enums.UserRoleViewer: 1,
}

type Api interface {
	ListServiceAlerts(ctx custom.Context, req *ServiceIdRequest) ([]*MemberAlertResponse, *service.Error)
	UpdateServiceAlert(ctx custom.Context, req *UpdateAlertRequest) (*MemberAlertResponse, *service.Error)
}

type AlertService struct {
	serviceRepo           *service_repo.ServiceRepo
	userRepo              *user_repo.UserRepo
	alertSubscriptionRepo *alert_subscription_repo.AlertSubscriptionRepo
}

var (
	alertService     *AlertService
	alertServiceOnce sync.Once
)

func GetAlertService() Api {
	alertServiceOnce.Do(func() {
		alertService = &AlertService{
			serviceRepo:           service_repo.GetServiceRepo(),
			userRepo:              user_repo.GetUserRepo(),
			alertSubscriptionRepo: alert_subscription_repo.GetAlertSubscriptionRepo(),
		}
	})
	return alertService
}

// ListServiceAlerts returns every member of the org with whether they get alerts for the service
func (s *AlertService) ListServiceAlerts(ctx custom.Context, req *ServiceIdRequest) ([]*MemberAlertResponse, *service.Error) {
	svc, repoErr := s.serviceRepo.GetByIdAndOrgId(ctx, req.Id, ctx.GetOrgId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	members, repoErr := s.userRepo.ListByOrgId(ctx, svc.OrgId)
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	subscriptions, repoErr := s.alertSubscriptionRepo.ListByServiceId(ctx, svc.Id, svc.OrgId)
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}

	byUserId := make(map[string]*models.AlertSubscription, len(subscriptions))
	for i := range subscriptions {
		byUserId[subscriptions[i].UserId] = &subscriptions[i]
	}
	names := make(map[string]string, len(members))
	for _, member := range members {
		names[member.Id] = member.Name
	}

	res := make([]*MemberAlertResponse, 0, len(members))
	for i := range members {
		res = append(res, newMemberAlertResponse(ctx.GetUser(), &members[i], byUserId[members[i].Id], names))
	}
	return res, nil
}

// UpdateServiceAlert turns a member's alerts for a service on or off. Members can change their
// own; a higher role can change a lower role's.
func (s *AlertService) UpdateServiceAlert(ctx custom.Context, req *UpdateAlertRequest) (*MemberAlertResponse, *service.Error) {
	svc, repoErr := s.serviceRepo.GetByIdAndOrgId(ctx, req.Id, ctx.GetOrgId())
	if repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	member, repoErr := s.userRepo.GetByIdAndOrgId(ctx, req.UserId, svc.OrgId)
	if repoErr != nil {
		if repoErr.Type == repo.NotFound {
			return nil, service.NewError("member not found", http.StatusNotFound)
		}
		return nil, service.HandleRepoErr(repoErr)
	}
	if !canChange(ctx.GetUser(), member) {
		return nil, service.NewError("you can only change your own alerts or those of a lower role", http.StatusForbidden)
	}

	subscription := &models.AlertSubscription{
		OrgId:     svc.OrgId,
		UserId:    member.Id,
		ServiceId: svc.Id,
		IsEnabled: *req.Enabled,
		UpdatedBy: ctx.GetUserId(),
	}
	if repoErr := s.alertSubscriptionRepo.Upsert(ctx, subscription); repoErr != nil {
		return nil, service.HandleRepoErr(repoErr)
	}
	names := map[string]string{ctx.GetUserId(): ctx.GetUser().Name}
	return newMemberAlertResponse(ctx.GetUser(), member, subscription, names), nil
}

// enabledByDefault tells whether a member gets alerts for a service they have no row for:
// owners do, editors and viewers do not
func enabledByDefault(role enums.UserRole) bool {
	return role == enums.UserRoleOwner
}

func canChange(caller *custom.UserProfileCtx, member *models.User) bool {
	return caller.UserID == member.Id || roleRank[caller.Role] > roleRank[member.Role]
}

func newMemberAlertResponse(caller *custom.UserProfileCtx, member *models.User,
	subscription *models.AlertSubscription, names map[string]string) *MemberAlertResponse {
	res := &MemberAlertResponse{
		UserId:    member.Id,
		Name:      member.Name,
		Role:      member.Role,
		Enabled:   enabledByDefault(member.Role),
		CanChange: canChange(caller, member),
	}
	if subscription != nil {
		res.Enabled = subscription.IsEnabled
		res.UpdatedAt = &subscription.UpdatedAt
		if name, ok := names[subscription.UpdatedBy]; ok {
			res.UpdatedByName = &name
		}
	}
	return res
}
