package archer

import (
	"context"
	"database/sql"
	"time"

	"github.com/dyaksa/archer/job"
	"github.com/dyaksa/archer/store"
)

type Tx interface {
	Schedule(ctx context.Context, id string, queueName string, arguments interface{}, options ...FnOptions) error
	Cancel(ctx context.Context, id string) error
	ScheduleNow(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*job.Job, error)
	Poll(ctx context.Context, queueName string) (*job.Job, error)
	Update(ctx context.Context, job job.Job) error
	RequeueTimeout(ctx context.Context, queueName string, timeout time.Time) error
}

type dbTx interface {
	Get(ctx context.Context, id string) (*job.Job, error)
	Poll(ctx context.Context, queueName string) (*job.Job, error)
	RequeueTimeout(ctx context.Context, queueName string, timeout time.Time) error
	Create(ctx context.Context, job job.Job) error
	Update(ctx context.Context, job job.Job) error
	Deschedule(ctx context.Context, id string) error
	ScheduleNow(ctx context.Context, id string) error
}

type transactionClient struct {
	tx dbTx
}

// Cancel implements Tx.
func (t *transactionClient) Cancel(ctx context.Context, id string) error {
	return t.tx.Deschedule(ctx, id)
}

// Get implements Tx.
func (t *transactionClient) Get(ctx context.Context, id string) (*job.Job, error) {
	return t.tx.Get(ctx, id)
}

// Schedule implements Tx.
func (t *transactionClient) Schedule(ctx context.Context, id string, queueName string, arguments interface{}, options ...FnOptions) error {
	var err error
	job := job.Job{
		ID:         id,
		QueueName:  queueName,
		Status:     job.StatusScheduled,
		ScheduleAt: time.Now(),
	}

	if job, err = job.SetArgs(arguments); err != nil {
		return err
	}

	for _, opt := range options {
		job = opt(job)
	}

	return t.tx.Create(ctx, job)
}

// ScheduleNow implements Tx.
func (t *transactionClient) ScheduleNow(ctx context.Context, id string) error {
	return t.tx.ScheduleNow(ctx, id)
}

// Poll implements Tx.
func (t *transactionClient) Poll(ctx context.Context, queueName string) (*job.Job, error) {
	return t.tx.Poll(ctx, queueName)
}

func (t *transactionClient) RequeueTimeout(ctx context.Context, queueName string, timeout time.Time) error {
	return t.tx.RequeueTimeout(ctx, queueName, timeout)
}

func (t *transactionClient) Update(ctx context.Context, job job.Job) error {
	return t.tx.Update(ctx, job)
}

func newTx(tx *sql.Tx, tableName string) Tx {
	return &transactionClient{tx: store.NewTx(tx, tableName)}
}
