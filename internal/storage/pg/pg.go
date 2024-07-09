package pg

import (
	"fmt"
	"log/slog"
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
  WHERE alpha2 = '%s'
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

func (pg *PgStorage) RegisterUser(user *models.User) (*models.Profile, error) {
	_, err := pg.db.Exec(fmt.Sprintf(`
    INSERT INTO users (login, email, country_fk, is_public, phone, image, password)
    VALUES ('%s', '%s', (
      SELECT id
      FROM countries
      WHERE alpha2 = '%s'
      LIMIT 1),
    %t, '%s', '%s', CAST('%s' AS BYTEA));`,
		user.Login, user.Email, user.CountryCode, user.IsPublic, user.Phone, user.Image, user.Password))
	if err != nil {
		return nil, err
	}

	return user.Profile(), nil
}

func (pg *PgStorage) User(login string) (*models.User, error) {
	row := pg.db.QueryRowx(fmt.Sprintf(`
	SELECT login, email, alpha2 AS CountryCode, is_public, phone, image, password
	FROM users
	INNER JOIN countries
	ON countries.id = country_fk
	WHERE login = '%s';`, login))
	if err := row.Err(); err != nil {
		return nil, err
	}
	// TODO: not found on nil value

	var u models.User
	if err := row.StructScan(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (pg *PgStorage) Profile(login string) (*models.Profile, error) {
	user, err := pg.User(login)
	if err != nil {
		return nil, err
	}

	return user.Profile(), nil
}

func (pg *PgStorage) Disconnect() error {
	return pg.db.Close()
}

// listString [a b c] -> 'a','b','c'
func listString(strs ...string) string {
	return "'" + strings.Join(strs, "','") + "'"
}
