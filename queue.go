package archer

import (
	"context"
	"database/sql"
	"time"

	"github.com/dyaksa/archer/job"
	"github.com/dyaksa/archer/store"
)

type Queue struct {
	store.WrapperTx
	tx   func(*sql.Tx) Tx
	now  func() time.Time
	name string
}

func NewQueue(db *sql.DB, name string, tableName string) *Queue {
	return &Queue{
		WrapperTx: *store.NewWrapperTx(db),
		now:       time.Now,
		name:      name,
		tx: func(tx *sql.Tx) Tx {
			return newTx(tx, tableName)
		},
	}
}

func (q *Queue) Poll(ctx context.Context) (*job.Job, error) {
	res, err := q.WrapTx(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		return q.tx(tx).Poll(ctx, q.name)
	})
	if err != nil {
		return nil, err
	}

	return res.(*job.Job), nil
}

func (q *Queue) RequeueTimeout(ctx context.Context, timeout time.Duration) error {
	_, err := q.WrapTx(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		return nil, q.tx(tx).RequeueTimeout(ctx, q.name, q.now().Add(-timeout))
	})

	return err
}
