package pg

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"solution/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PgStorage struct {
	db *sqlx.DB
}

func New(connString string, logger *slog.Logger) (*PgStorage, error) {
	db, err := sqlx.Connect("pgx", connString)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS countries (
		id SERIAL PRIMARY KEY,
		name TEXT UNIQUE,
		alpha2 TEXT,
		alpha3 TEXT,
		region TEXT
	)`)
	if err != nil {
		logger.Error("could not create table countries", slog.String("err", err.Error()))
	}
	err = addCountries(db)
	if err != nil {
		logger.Error("could not insert countries", slog.String("err", err.Error()))
	}

	return &PgStorage{
		db: db,
	}, nil
}

func (pg *PgStorage) Countries(regions ...string) (countries []models.Country, err error) {
	if len(regions) == 0 {
		return pg.AllCountries()
	}

	countries = make([]models.Country, 0, 256)
	err = pg.db.Select(&countries, fmt.Sprintf(`
	SELECT name, alpha2, alpha3, region
	FROM countries
	WHERE region IN(%s)
	ORDER BY alpha2 ASC`, listString(regions...)))
	return
}

func (pg *PgStorage) AllCountries() (countries []models.Country, err error) {
	countries = make([]models.Country, 0, 256)

	err = pg.db.Select(&countries, `SELECT name, alpha2, alpha3, region FROM countries ORDER BY alpha2`)
	return countries, err
}

func (pg *PgStorage) Disconnect() error {
	return pg.db.Close()
}

// listString [a b c] -> 'a','b','c'
func listString(strs ...string) string {
	return "'" + strings.Join(strs, "','") + "'"
}

func addCountries(db *sqlx.DB) error {
	row := db.QueryRow(`SELECT COUNT(name) AS count FROM countries`)
	var rowsCount struct {
		Count int
	}
	row.Scan(&rowsCount)
	if rowsCount.Count == 249 {
		return nil
	}

	b, err := os.ReadFile("./resources/countryinsert.txt")
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf(`INSERT INTO countries (name, alpha2, alpha3, region) VALUES %s;`, string(b)))
	return err
}
