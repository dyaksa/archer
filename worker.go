package archer

import (
	"context"

	"github.com/dyaksa/archer/job"
)

type Handler interface {
	Handle(ctx context.Context, job job.Job) error
}

type WorkerFn func(ctx context.Context, job job.Job) (any, error)

type Worker interface {
	Execute(ctx context.Context, job job.Job) (any, error)
	OnFailure(ctx context.Context, job job.Job) error
}

type fnWorker struct {
	fn WorkerFn
}

func (f *fnWorker) Execute(ctx context.Context, job job.Job) (any, error) {
	return f.fn(ctx, job)
}

func (f *fnWorker) OnFailure(ctx context.Context, job job.Job) error {
	return nil
}
