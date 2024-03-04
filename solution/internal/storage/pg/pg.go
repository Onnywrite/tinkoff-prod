package pg

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"solution/internal/models"
	"solution/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PgStorage struct {
	db     *sqlx.DB
	logger *slog.Logger
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
		db:     db,
		logger: logger,
	}, nil
}

func (pg *PgStorage) Countries(regions ...string) ([]models.Country, error) {
	const op = "pg.PgStorage_Countries"

	if len(regions) == 0 {
		return pg.AllCountries()
	}

	countries := make([]models.Country, 0, 256)
	err := pg.db.Select(&countries, fmt.Sprintf(`
	SELECT name, alpha2, alpha3, region
	FROM countries
	WHERE region IN(%s)
	ORDER BY alpha2 ASC`, listString(regions...)))
	if err != nil {
		pg.logger.Error("error while selecting countries", slog.String("err", err.Error()), slog.String("op", op))
		return nil, storage.ErrInternal
	}

	if len(countries) == 0 {
		pg.logger.Error("found 0 countries", slog.Any("regions", regions), slog.String("op", op))
		return nil, storage.ErrCountriesNotFound
	}

	return countries, nil
}

func (pg *PgStorage) AllCountries() ([]models.Country, error) {
	const op = "pg.PgStorage_AllCountries"

	countries := make([]models.Country, 0, 256)

	err := pg.db.Select(&countries, `SELECT name, alpha2, alpha3, region FROM countries ORDER BY alpha2`)
	if err != nil {
		pg.logger.Error("error while selecting all countries", slog.String("err", err.Error()), slog.String("op", op))
		return nil, storage.ErrInternal
	}

	return countries, nil
}

func (pg *PgStorage) Country(alpha2 string) (models.Country, error) {
	const op = "pg.PgStorage_Country"

	row := pg.db.QueryRowx(fmt.Sprintf(`
  SELECT name, alpha2, alpha3, region
  FROM countries
  WHERE alpha2 = %s
  `, alpha2))
	if row.Err() != nil {
		pg.logger.Error("error while selecting a country by alpha2", slog.String("err", row.Err().Error()), slog.String("op", op))
		return models.Country{}, storage.ErrInternal
	}

	var c models.Country
	err := row.StructScan(&c)
	if err != nil {
		pg.logger.Error("error while getting a country by alpha2", slog.String("err", err.Error()), slog.String("op", op))
		return models.Country{}, storage.ErrCountryNotFound
	}

	return c, nil

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
