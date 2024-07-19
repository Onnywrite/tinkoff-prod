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
)

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
