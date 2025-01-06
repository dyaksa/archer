package store

import (
	"context"
	"database/sql"
)

func exec(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) error {
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}
