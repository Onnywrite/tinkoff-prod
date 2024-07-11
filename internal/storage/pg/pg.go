package pg

import (
	"context"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"

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

func (pg *PgStorage) Countries(ctx context.Context, regions ...string) ([]models.Country, error) {
	const op = "pg.PgStorage_Countries"

	if len(regions) == 0 {
		return pg.AllCountries(ctx)
	}

	countries := make([]models.Country, 0, 256)

	stmt, err := pg.db.PreparexContext(ctx, `
		SELECT id, name, alpha2, alpha3, region
		FROM countries
		WHERE region IN($1)
		ORDER BY alpha2`,
	)
	if err != nil {
		return nil, err
	}

	err = stmt.SelectContext(ctx, &countries, regions)
	if err != nil {
		return nil, storage.ErrInternal
	}

	if len(countries) == 0 {
		return nil, storage.ErrCountriesNotFound
	}

	return countries, nil
}

func (pg *PgStorage) AllCountries(ctx context.Context) ([]models.Country, error) {
	const op = "pg.PgStorage_AllCountries"

	countries := make([]models.Country, 0, 256)

	err := pg.db.SelectContext(ctx, &countries, `
		SELECT id, name, alpha2, alpha3, region
		FROM countries
		ORDER BY alpha2`,
	)
	if err != nil {
		return nil, storage.ErrInternal
	}

	return countries, nil
}

func (pg *PgStorage) Country(ctx context.Context, alpha2 string) (c models.Country, err error) {
	const op = "pg.PgStorage_Country"

	stmt, err := pg.db.PreparexContext(ctx, `
  		SELECT id, name, alpha2, alpha3, region
  		FROM countries
  		WHERE alpha2 = $1`,
	)
	if err != nil {
		return models.Country{}, err
	}

	err = stmt.GetContext(ctx, &c, alpha2)
	if err != nil {
		return models.Country{}, storage.ErrInternal
	}

	return c, nil
}

func (pg *PgStorage) SaveUser(ctx context.Context, user *models.User) (*models.User, error) {
	const op = "pg.PgStorage_SaveUser"

	type returning struct {
		Id           uint64    `db:"id"`
		Name         string    `db:"name"`
		Lastname     string    `db:"lastname"`
		Email        string    `db:"email"`
		IsPublic     bool      `db:"is_public"`
		Image        string    `db:"image"`
		Birthday     time.Time `db:"birthday"`
		PasswordHash string    `db:"password"`
		CountryId    uint64    `db:"c_id"`
		CountryName  string    `db:"c_name"`
		Alpha2       string    `db:"alpha2"`
		Alpha3       string    `db:"alpha3"`
		Region       string    `db:"region"`
	}

	stmt, err := pg.db.PreparexContext(ctx, `
    	WITH u AS (
			INSERT INTO users (name, lastname, email, country_fk, is_public, image, password, birthday)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING *
		)
		SELECT u.id, u.name, u.lastname, u.email, u.is_public, u.image, u.password, u.birthday,
			   countries.id AS c_id, countries.name AS c_name, countries.alpha2, countries.alpha3, countries.region
		FROM u
		JOIN countries ON countries.id = country_fk`,
	)
	if err != nil {
		return nil, err
	}

	var saved returning
	err = stmt.QueryRowxContext(ctx, user.Name, user.Lastname, user.Email, user.Country.Id, user.IsPublic, user.Image, user.PasswordHash, user.Birthday).
		StructScan(&saved)
	if err != nil {
		return nil, err
	}

	return &models.User{
		Id:           saved.Id,
		Name:         saved.Name,
		Lastname:     saved.Lastname,
		Email:        saved.Email,
		IsPublic:     saved.IsPublic,
		Image:        saved.Image,
		Birthday:     saved.Birthday,
		PasswordHash: saved.PasswordHash,
		Country: models.Country{
			Id:     saved.CountryId,
			Name:   saved.CountryName,
			Alpha2: saved.Alpha2,
			Alpha3: saved.Alpha3,
			Region: saved.Region,
		},
	}, nil
}

func (pg *PgStorage) UserByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "pg.PgStorage_UserByEmail"

	type returning struct {
		Id           uint64    `db:"id"`
		Name         string    `db:"name"`
		Lastname     string    `db:"lastname"`
		Email        string    `db:"email"`
		IsPublic     bool      `db:"is_public"`
		Image        string    `db:"image"`
		Birthday     time.Time `db:"birthday"`
		PasswordHash string    `db:"password"`
		CountryId    uint64    `db:"c_id"`
		CountryName  string    `db:"c_name"`
		Alpha2       string    `db:"alpha2"`
		Alpha3       string    `db:"alpha3"`
		Region       string    `db:"region"`
	}

	stmt, err := pg.db.PreparexContext(ctx, `
		SELECT users.id, users.name, users.lastname, users.email, users.is_public, users.image, users.password, users.birthday,
			   countries.id AS c_id, countries.name AS c_name, countries.alpha2, countries.alpha3, countries.region
		FROM users
		JOIN countries
		ON countries.id = country_fk
		WHERE email = $1`,
	)
	if err != nil {
		return nil, err
	}

	var u returning
	err = stmt.GetContext(ctx, &u, email)
	if err != nil {
		return nil, err
	}

	return &models.User{
		Id:           u.Id,
		Name:         u.Name,
		Lastname:     u.Lastname,
		Email:        u.Email,
		IsPublic:     u.IsPublic,
		Image:        u.Image,
		Birthday:     u.Birthday,
		PasswordHash: u.PasswordHash,
		Country: models.Country{
			Id:     u.CountryId,
			Name:   u.CountryName,
			Alpha2: u.Alpha2,
			Alpha3: u.Alpha3,
			Region: u.Region,
		},
	}, nil
}

func (pg *PgStorage) Disconnect() error {
	return pg.db.Close()
}
