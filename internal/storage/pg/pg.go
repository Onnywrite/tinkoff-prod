package pg

import (
	"github.com/Onnywrite/tinkoff-prod/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// copied from https://github.com/jackc/pgerrcode/blob/master/errcode.go
const (
	notNullViolation    = "23502"
	foreignKeyViolation = "23503"
	uniqueViolation     = "23505"
	checkViolation      = "23514"
)

var pgerrToErr = map[string]error{
	notNullViolation:    storage.ErrNotNullConstraint,
	foreignKeyViolation: storage.ErrForeignKeyConstraint,
	uniqueViolation:     storage.ErrUniqueConstraint,
	checkViolation:      storage.ErrCheckConstraint,
}

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
