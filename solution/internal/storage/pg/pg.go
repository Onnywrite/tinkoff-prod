package pg

import (
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

	db.Exec(`
	CREATE TABLE IF NOT EXISTS (
		id SERIAL PRIMARY KEY,
		name TEXT,
		alpha2 TEXT,
		alpha3 TEXT,
		region TEXT
	)`)

	return &PgStorage{
		db: db,
	}, nil
}

func (pg *PgStorage) Disconnect() error {
	return pg.db.Close()
}
