package async_middleware

import (
	"context"

	"github.com/riverqueue/river"

	"pingwise/common/custom"
)

type systemWorker[T river.JobArgs] struct {
	river.WorkerDefaults[T]
	fn func(ctx custom.Context, job *river.Job[T]) error
}

func (w *systemWorker[T]) Work(ctx context.Context, job *river.Job[T]) error {
	return w.fn(custom.NewCustomContext(ctx), job)
}

// WithSystemWorker turns a function into a river worker that gets a custom.Context.
// System workers run without a logged in user.
func WithSystemWorker[T river.JobArgs](fn func(ctx custom.Context, job *river.Job[T]) error) river.Worker[T] {
	return &systemWorker[T]{fn: fn}
}
