package postgres

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGErrorClassification int

const (
	NonRetriable PGErrorClassification = iota
	Retriable
)

func classifyPGError(err error) PGErrorClassification {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgerrcode.IsConnectionException(pgErr.Code) {
			return Retriable
		}
	}
	return NonRetriable
}
