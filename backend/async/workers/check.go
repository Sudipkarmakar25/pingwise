package workers

import (
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"pingwise/async/args"
	"pingwise/async/async_middleware"
	"pingwise/common/custom"
	"pingwise/common/logger"
	"pingwise/internal/service/check_service"
)

// scheduleChecksWorker claims the services that are due and queues one check job for each
func scheduleChecksWorker(ctx custom.Context, job *river.Job[args.ScheduleChecksArgs]) error {
	dueServices, serviceErr := check_service.GetCheckService().ClaimDueServices(ctx)
	if serviceErr != nil {
		return serviceErr
	}
	if len(dueServices) == 0 {
		return nil
	}

	params := make([]river.InsertManyParams, 0, len(dueServices))
	for _, svc := range dueServices {
		params = append(params, river.InsertManyParams{Args: args.CheckServiceArgs{ServiceId: svc.Id, OrgId: svc.OrgId}})
	}
	if _, err := river.ClientFromContext[pgx.Tx](ctx).InsertMany(ctx, params); err != nil {
		// the claimed services are picked up again once their interval passes
		logger.Errorf("failed to queue %d checks : %v", len(params), err)
		return err
	}
	logger.Infof("queued %d checks", len(params))
	return nil
}

// checkServiceWorker runs one check and, when the service went down or came back up, queues an alert
func checkServiceWorker(ctx custom.Context, job *river.Job[args.CheckServiceArgs]) error {
	change, serviceErr := check_service.GetCheckService().RunCheck(ctx, job.Args.ServiceId, job.Args.OrgId)
	if serviceErr != nil {
		return serviceErr
	}
	if change == nil {
		return nil
	}
	alert := args.SendAlertArgs{
		ServiceId:           change.ServiceId,
		OrgId:               change.OrgId,
		Status:              string(change.Status),
		CheckedAt:           change.CheckedAt,
		StatusCode:          change.StatusCode,
		Error:               change.Error,
		ConsecutiveFailures: change.ConsecutiveFailures,
	}
	if _, err := river.ClientFromContext[pgx.Tx](ctx).Insert(ctx, alert, nil); err != nil {
		// the check itself is saved; only the alert is lost
		logger.Errorf("failed to queue alert for service %s : %v", change.ServiceId, err)
		return nil
	}
	logger.Infof("service %s is now %s, alert queued", change.ServiceId, change.Status)
	return nil
}

func NewScheduleChecksWorker() river.Worker[args.ScheduleChecksArgs] {
	return async_middleware.WithSystemWorker(scheduleChecksWorker)
}

func NewCheckServiceWorker() river.Worker[args.CheckServiceArgs] {
	return async_middleware.WithSystemWorker(checkServiceWorker)
}
