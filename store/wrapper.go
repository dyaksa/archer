package store

import (
	"context"
	"database/sql"
)

type WrapperTx struct {
	db *sql.DB
}

func NewWrapperTx(db *sql.DB) *WrapperTx {
	return &WrapperTx{db: db}
}

func (w *WrapperTx) WrapTx(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) (any, error)) (any, error) {
	tx, err := w.db.Begin()
	if err != nil {
		return nil, err
	}

	res, err := fn(ctx, tx)
	switch err {
	case nil:
	case sql.ErrNoRows:
		_ = tx.Commit()
		return res, nil
	default:
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return res, nil
}
