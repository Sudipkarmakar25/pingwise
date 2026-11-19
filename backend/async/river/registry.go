package river

import (
	"time"

	"github.com/riverqueue/river"

	"pingwise/async/args"
	"pingwise/async/workers"
)

// Developer has to make sure that every task is idempotent

func RegisterWorkerTasks() *river.Workers {
	riverWorkers := river.NewWorkers()

	river.AddWorker[args.ScheduleChecksArgs](riverWorkers, workers.NewScheduleChecksWorker())
	river.AddWorker[args.CheckServiceArgs](riverWorkers, workers.NewCheckServiceWorker())
	river.AddWorker[args.SendAlertArgs](riverWorkers, workers.NewSendAlertWorker())
	river.AddWorker[args.SendEmailArgs](riverWorkers, workers.NewSendEmailWorker())

	return riverWorkers
}

func RegisterPeriodicJobs() []*river.PeriodicJob {
	return []*river.PeriodicJob{
		river.NewPeriodicJob(
			startOfEveryMinute{},
			func() (river.JobArgs, *river.InsertOpts) {
				return args.ScheduleChecksArgs{}, nil
			},
			&river.PeriodicJobOpts{ID: "schedule_checks", RunOnStart: true},
		),
	}
}

// startOfEveryMinute fires at :00 of every minute. Ticks lined up on the minute
// match the minute-truncated last_scheduled_at the scheduler compares against.
type startOfEveryMinute struct{}

func (startOfEveryMinute) Next(current time.Time) time.Time {
	return current.Truncate(time.Minute).Add(time.Minute)
}
