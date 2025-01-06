package archer

import (
	"context"
	"database/sql"
	"time"

	"github.com/dyaksa/archer/job"
	"github.com/dyaksa/archer/store"
)

type mutate interface {
	Update(ctx context.Context, job job.Job) error
}

type Mutate struct {
	store.WrapperTx
	tx func(*sql.Tx) Tx
}

func newMutate(db *sql.DB) *Mutate {
	return &Mutate{
		WrapperTx: *store.NewWrapperTx(db),
		tx: func(tx *sql.Tx) Tx {
			return newTx(tx)
		},
	}
}

func (m *Mutate) Update(ctx context.Context, j job.Job) error {
	_, err := m.WrapTx(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		return nil, m.tx(tx).Update(ctx, j)
	})
	return err
}

type handler struct {
	worker Worker
	mutate mutate
}

func newHandler(w Worker, mutate mutate) *handler {
	return &handler{
		worker: w,
		mutate: mutate,
	}
}

func (h *handler) Handle(ctx context.Context, job job.Job) error {
	res, err := h.worker.Execute(ctx, job)
	if err != nil {
		return h.failure(ctx, job, err)
	}

	return h.success(ctx, job, res)
}

func (h *handler) failure(ctx context.Context, j job.Job, err error) error {
	j = j.SetLastError(err)

	if j.ShouldRetry() {
		retryAt := time.Now().Add(j.RetryInterval)
		j = j.ScheduleRetry(retryAt)
		return h.mutate.Update(ctx, j)
	}

	j = j.SetStatus(job.StatusFailed)

	if err := h.mutate.Update(ctx, j); err != nil {
		return err
	}

	return h.worker.OnFailure(ctx, j)
}

func (h *handler) success(ctx context.Context, j job.Job, res any) error {
	j = j.SetStatus(job.StatusCompleted)

	var err error
	if j, err = j.SetResult(res); err != nil {
		j = j.SetLastError(err)
	}

	return h.mutate.Update(ctx, j)
}
