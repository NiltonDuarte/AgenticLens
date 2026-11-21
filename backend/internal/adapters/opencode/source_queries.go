package opencode

import (
	"context"
	"database/sql"
)

// sourceQueries keeps exact snapshots on their transaction and ordinary reads
// on the selected pool. The call sites supply typed SQL parameters.
type sourceQueries struct {
	db *sql.DB
	tx *sql.Tx
}

// Keep expensive source reads under modernc's QueryContext interrupt guard.
// The driver drops that guard before later Rows.Next calls.
func materializeSourceRows(ctx context.Context, statement, orderBy string) string {
	if ctx.Done() == nil {
		return statement
	}
	query := "WITH source_rows AS MATERIALIZED (" + statement + ") SELECT * FROM source_rows"
	if orderBy != "" {
		query += " ORDER BY " + orderBy
	}
	return query
}

func (s sourceQueries) query(ctx context.Context, statement string) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.QueryContext(ctx, statement)
	}
	return s.db.QueryContext(ctx, statement)
}

func (s sourceQueries) querySession(ctx context.Context, statement, id string) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.QueryContext(ctx, statement, id)
	}
	return s.db.QueryContext(ctx, statement, id)
}

func (s sourceQueries) queryActivity(ctx context.Context, statement string, cutoff int64) (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.QueryContext(ctx, statement, cutoff, cutoff, cutoff)
	}
	return s.db.QueryContext(ctx, statement, cutoff, cutoff, cutoff)
}

func (s sourceQueries) queryRow(ctx context.Context, statement string) *sql.Row {
	if s.tx != nil {
		return s.tx.QueryRowContext(ctx, statement)
	}
	return s.db.QueryRowContext(ctx, statement)
}

func (s sourceQueries) querySessionRow(ctx context.Context, statement, id string) *sql.Row {
	if s.tx != nil {
		return s.tx.QueryRowContext(ctx, statement, id)
	}
	return s.db.QueryRowContext(ctx, statement, id)
}
