package river

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"

	"pingwise/common/logger"
)

func InitRiverClient(ctx context.Context, dbDSN string, checkWorkers int) (*river.Client[pgx.Tx], error) {
	dbPool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	riverClient, err := river.NewClient[pgx.Tx](riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: checkWorkers},
		},
		ErrorHandler:                &DefaultErrorHandler{},
		JobTimeout:                  2 * time.Minute,
		MaxAttempts:                 1,
		PeriodicJobs:                RegisterPeriodicJobs(),
		CompletedJobRetentionPeriod: time.Hour, // results live in check_results, so finished jobs need not be kept
		DiscardedJobRetentionPeriod: 7 * 24 * time.Hour,
		Workers:                     RegisterWorkerTasks(),
		Logger:                      slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn})),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create River client: %w", err)
	}
	return riverClient, nil
}

type DefaultErrorHandler struct{}

func (h *DefaultErrorHandler) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	logger.Errorf("Error processing for job_id: %v, kind: %v, err: %v", job.ID, job.Kind, err.Error())
	return nil
}

func (h *DefaultErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	logger.Errorf("Panic processing for job_id: %v, kind: %v, panic: %v\n%s", job.ID, job.Kind, panicVal, trace)
	return nil
}
