package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"

	"github.com/jackc/pgx/v5/pgconn"
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

func (pg *PgStorage) Countries(ctx context.Context, regions ...string) ([]models.Country, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "pg.PgStorage.Countries").With("regions", regions)

	if len(regions) == 0 {
		return pg.AllCountries(ctx)
	}

	query, args, err := sqlx.In(`
		SELECT id, name, alpha2, alpha3, region
		FROM countries
		WHERE region IN(?)
		ORDER BY alpha2`,
		regions,
	)
	if err != nil {
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	stmt, err := pg.db.PreparexContext(ctx, pg.db.Rebind(query))
	if err != nil {
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	countries := make([]models.Country, 0, 256)

	err = stmt.SelectContext(ctx, &countries, args...)
	if err != nil {
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	if len(countries) == 0 {
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeNotFound, storage.ErrNoRows)
	}

	return countries, nil
}

func (pg *PgStorage) AllCountries(ctx context.Context) ([]models.Country, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "pg.PgStorage.AllCountries")

	countries := make([]models.Country, 0, 256)

	err := pg.db.SelectContext(ctx, &countries, `
		SELECT id, name, alpha2, alpha3, region
		FROM countries
		ORDER BY alpha2`,
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeNotFound, storage.ErrNoRows)
	case err != nil:
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	return countries, nil
}

func (pg *PgStorage) Country(ctx context.Context, alpha2 string) (models.Country, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "pg.PgStorage.Country").With("alpha2", alpha2)

	stmt, err := pg.db.PreparexContext(ctx, `
  		SELECT id, name, alpha2, alpha3, region
  		FROM countries
  		WHERE alpha2 = $1`,
	)
	if err != nil {
		return models.Country{}, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	var c models.Country
	err = stmt.GetContext(ctx, &c, alpha2)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.Country{}, ero.New(logCtx.With("error", err).Build(), ero.CodeNotFound, storage.ErrNoRows)
	case err != nil:
		return models.Country{}, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	return c, nil
}

func (pg *PgStorage) SaveUser(ctx context.Context, user *models.User) (*models.User, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "pg.PgStorage.SaveUser").With("user_email", user.Email)

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
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	var saved returning
	err = stmt.GetContext(ctx, &saved,
		user.Name, user.Lastname, user.Email, user.Country.Id,
		user.IsPublic, user.Image, user.PasswordHash, user.Birthday)

	// TODO: refactor ASAP
	if err != nil {
		pgErr := &pgconn.PgError{}
		var strerr string
		if errors.As(err, &pgErr) {
			strerr = pgErr.Code
		} else {
			strerr = err.Error()
		}
		doneErr, ok := pgerrToErr[strerr]
		if !ok {
			doneErr = storage.ErrInternal
		}
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, doneErr)
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

func (pg *PgStorage) UserByEmail(ctx context.Context, email string) (*models.User, ero.Error) {
	return pg.userBy(ctx, "users.email = $1", email)
}

func (pg *PgStorage) UserById(ctx context.Context, id uint64) (*models.User, ero.Error) {
	return pg.userBy(ctx, "users.id = $1", id)
}

func (pg *PgStorage) userBy(ctx context.Context, where string, args ...any) (*models.User, ero.Error) {
	logCtx := erolog.NewContextBuilder().WithParent(ctx).With("op", "pg.PgStorage.userBy").With("args", args)

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

	stmt, err := pg.db.PreparexContext(ctx, fmt.Sprintf(`
		SELECT users.id, users.name, users.lastname, users.email, users.is_public, users.image, users.password, users.birthday,
			   countries.id AS c_id, countries.name AS c_name, countries.alpha2, countries.alpha3, countries.region
		FROM users
		JOIN countries
		ON countries.id = country_fk
		WHERE %s`, where),
	)
	if err != nil {
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	var u returning
	err = stmt.GetContext(ctx, &u, args...)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeNotFound, storage.ErrNoRows)
	case err != nil:
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
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

func (pg *PgStorage) SavePost(ctx context.Context, post *models.Post) (uint64, ero.Error) {
	logCtx := erolog.NewContextBuilder().WithParent(ctx).With("op", "pg.PgStorage.userBy").With("post_author_id", post.Author.Id)

	stmt, err := pg.db.PreparexContext(ctx, `
		INSERT INTO posts (author_fk, content, images_urls)
		VALUES ($1, $2, $3)
		RETURNING id`,
	)
	if err != nil {
		return 0, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	var id uint64
	err = stmt.GetContext(ctx, &id, post.Author.Id, post.Content, post.ImagesUrls)

	// TODO: refactor ASAP
	if err != nil {
		pgErr := &pgconn.PgError{}
		var strerr string
		if errors.As(err, &pgErr) {
			strerr = pgErr.Code
		} else {
			strerr = err.Error()
		}
		doneErr, ok := pgerrToErr[strerr]
		if !ok {
			doneErr = storage.ErrInternal
		}
		return 0, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, doneErr)
	}

	return id, nil
}

func (pg *PgStorage) Disconnect() error {
	return pg.db.Close()
}
