package archer

import (
	"context"
	"database/sql"
	"time"

	"github.com/dyaksa/archer/job"
	"github.com/dyaksa/archer/store"
)

// mutate is an interface that defines a method for updating a job.
// The Update method takes a context and a job as parameters and returns an error if the update fails.
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

// Update updates the given job in the database within a transaction.
// It wraps the update operation in a transaction context and ensures
// that the transaction is properly managed.
//
// Parameters:
//   - ctx: The context for the update operation.
//   - j: The job to be updated.
//
// Returns:
//   - error: An error if the update operation fails, otherwise nil.
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

// Handle processes a job by executing it with the worker and handling the result.
// If the execution fails, it calls the failure handler with the error.
// If the execution succeeds, it calls the success handler with the result.
//
// Parameters:
//
//	ctx - The context for controlling cancellation and deadlines.
//	job - The job to be processed.
//
// Returns:
//
//	An error if the job processing fails, otherwise nil.
func (h *handler) Handle(ctx context.Context, job job.Job) error {
	res, err := h.worker.Execute(ctx, job)
	if err != nil {
		return h.failure(ctx, job, err)
	}

	return h.success(ctx, job, res)
}

// failure handles the failure of a job by updating its status and scheduling a retry if applicable.
// If the job should be retried, it schedules the retry and updates the job in the datastore.
// If the job should not be retried, it sets the job status to failed and updates the job in the datastore.
// Finally, it calls the worker's OnFailure method to handle any additional failure logic.
//
// Parameters:
//
//	ctx - the context for the operation
//	j - the job that has failed
//	err - the error that caused the job to fail
//
// Returns:
//
//	an error if there was an issue updating the job or handling the failure, otherwise nil
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

// success updates the job status to completed, sets the result of the job,
// and updates the job in the database. If setting the result fails, it sets
// the last error on the job.
//
// Parameters:
//
//	ctx - the context for the operation
//	j - the job to update
//	res - the result to set on the job
func (h *handler) success(ctx context.Context, j job.Job, res any) error {
	j = j.SetStatus(job.StatusCompleted)

	var err error
	if j, err = j.SetResult(res); err != nil {
		j = j.SetLastError(err)
	}

	return h.mutate.Update(ctx, j)
}
