package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dyaksa/archer/job"
)

func queryJob(ctx context.Context, tx *sql.Tx, query string, args ...any) (*job.Job, error) {
	e := new(entity)
	err := tx.QueryRowContext(ctx, query, args...).Scan(e.ScanDestinations()...)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return &job.Job{}, job.ErrorJobNotFound
	case err != nil:
		return &job.Job{}, err
	default:
		return e.To(), nil
	}
}

func queryJobs(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]*job.Job, error) {
	entities := []*job.Job{}
	rows, err := tx.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		e := new(entity)
		if err := rows.Scan(e.ScanDestinations()...); err != nil {
			return nil, err
		}
		entities = append(entities, e.To())
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entities, nil
}
