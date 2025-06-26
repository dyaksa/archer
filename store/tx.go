package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/dyaksa/archer/job"
)

type TxStore interface {
	Search(ctx context.Context, limit int, offset int, search string) ([]*job.Job, error)
	Get(ctx context.Context, id string) (*job.Job, error)
	Update(ctx context.Context, job job.Job) error
	Create(ctx context.Context, job job.Job) error
	Deschedule(ctx context.Context, id string) error
	ScheduleNow(ctx context.Context, id string) error
	Poll(ctx context.Context, queueName string) (*job.Job, error)
	RequeueTimeout(ctx context.Context, queueName string, timeout time.Time) error
	Commit() error
}

type Tx struct {
	*sql.Tx
	tableName string
}

func NewTx(tx *sql.Tx, tableName string) TxStore {
	return &Tx{tx, tableName}
}

func (t *Tx) Search(ctx context.Context, limit int, offset int, search string) ([]*job.Job, error) {
	if search != "" {
		return queryJobs(ctx, t.Tx, `SELECT `+entryFields+`
		FROM `+t.tableName+`
		WHERE id LIKE '%' || $1 || '%' 
		ORDER BY scheduled_at DESC 
		LIMIT $2 OFFSET $3`, search, limit, offset)
	}

	return queryJobs(ctx, t.Tx, `SELECT `+entryFields+`
	FROM `+t.tableName+`
	ORDER BY scheduled_at DESC 
	LIMIT $1 OFFSET $2`, limit, offset)
}

func (t *Tx) Get(ctx context.Context, id string) (*job.Job, error) {
	return queryJob(ctx, t.Tx, `SELECT `+entryFields+`
	FROM `+t.tableName+` 
	WHERE id = $1`, id)
}

func (t *Tx) Update(ctx context.Context, job job.Job) error {
	return exec(ctx, t.Tx, `UPDATE `+t.tableName+`
	SET
		status=$1, 
		result=$2, 
		last_error=$3, 
		retry_count=$4,
		scheduled_at=$5,
		updated_at=now()
	WHERE id = $6`,
		job.Status,
		job.Result,
		job.LastError,
		job.RetryCount,
		job.ScheduleAt,
		job.ID,
	)
}

func (t *Tx) Create(ctx context.Context, job job.Job) error {
	return exec(ctx, t.Tx, `INSERT INTO `+t.tableName+` (id, queue_name, status, arguments, max_retry, retry_interval, scheduled_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		job.ID, job.QueueName, job.Status, job.Arguments, job.MaxRetry, job.RetryInterval, job.ScheduleAt)
}

func (t *Tx) Deschedule(ctx context.Context, id string) error {
	return exec(ctx, t.Tx, `UPDATE `+t.tableName+` 
	SET 
		updated_at=now(), 
		status=$1 
	WHERE 
		id = $2 AND 
		status = $3`, job.StatusCanceled, id, job.StatusScheduled)
}

func (t *Tx) ScheduleNow(ctx context.Context, id string) error {
	return exec(ctx, t.Tx, `UPDATE `+t.tableName+` 
	SET 
		updated_at=now(), 
		scheduled_at=now(), 
		status=$1 
	WHERE 
		id = $2`, job.StatusScheduled, id)
}

func (t *Tx) Poll(ctx context.Context, queueName string) (*job.Job, error) {
	query := `UPDATE ` + t.tableName + `
		SET 
			status=$1, 
			started_at=now(),
			updated_at=now()
		WHERE 
			id = (
				SELECT id
				FROM ` + t.tableName + ` 
				WHERE status = $2
					AND scheduled_at <= now()
					AND queue_name = $3
				ORDER BY scheduled_at ASC 
				FOR UPDATE SKIP LOCKED
				LIMIT 1 
			)
		RETURNING ` + entryFields

	return queryJob(ctx, t.Tx, query, job.StatusInitialized, job.StatusScheduled, queueName)
}

func (t *Tx) RequeueTimeout(ctx context.Context, queueName string, timeout time.Time) error {
	return exec(ctx, t.Tx, `UPDATE `+t.tableName+` 
	SET 
		status=$1,
		started_at=null,
		retry_count=retry_count+1,
		updated_at=now()
	WHERE 
		started_at < $2 AND 
		status = $3 AND
		queue_name = $4`, job.StatusScheduled, timeout, job.StatusInitialized, queueName)
}
