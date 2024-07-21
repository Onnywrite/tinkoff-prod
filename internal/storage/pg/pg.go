package pg

import (
	"database/sql"
	"errors"

	"github.com/Onnywrite/tinkoff-prod/internal/storage"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PgStorage struct {
	db *sqlx.DB
}

func New(connString string) (*PgStorage, error) {
	db, err := sqlx.Connect("pgx", connString)
	if err != nil {
		return nil, err
	}

	return &PgStorage{
		db: db,
	}, nil
}

func (pg *PgStorage) Disconnect() error {
	return pg.db.Close()
}

// copied from https://github.com/jackc/pgerrcode/blob/master/errcode.go
const (
	notNullViolation    = "23502"
	foreignKeyViolation = "23503"
	uniqueViolation     = "23505"
	checkViolation      = "23514"
)

var errorsMap = map[string]error{
	notNullViolation:      storage.ErrNotNullConstraint,
	foreignKeyViolation:   storage.ErrForeignKeyConstraint,
	uniqueViolation:       storage.ErrUniqueConstraint,
	checkViolation:        storage.ErrCheckConstraint,
	sql.ErrNoRows.Error(): storage.ErrNoRows,
}

func getError(err error) error {
	pgErr := &pgconn.PgError{}
	var strerr string
	if errors.As(err, &pgErr) {
		strerr = pgErr.Code
	} else {
		strerr = err.Error()
	}
	doneErr, ok := errorsMap[strerr]
	if !ok {
		doneErr = storage.ErrInternal
	}
	return doneErr
}
