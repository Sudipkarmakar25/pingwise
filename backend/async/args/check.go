package args

import (
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// ScheduleChecksArgs is the periodic job that runs every minute and queues a
// CheckServiceArgs job for every service that is due.
type ScheduleChecksArgs struct{}

func (ScheduleChecksArgs) Kind() string {
	return "schedule_checks"
}

func (ScheduleChecksArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{MaxAttempts: 1}
}

// CheckServiceArgs runs one health check. It carries only ids; the worker reads the
// current settings from the services table, so an edit applies on the next check.
type CheckServiceArgs struct {
	ServiceId string `json:"service_id" river:"unique"`
	OrgId     string `json:"org_id"`
}

func (CheckServiceArgs) Kind() string {
	return "check_service"
}

// InsertOpts: a failed check is a result, not a job failure, so there are no retries.
// Only one check per service can be waiting or running at a time.
func (CheckServiceArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 1,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
	}
}
