package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

func (pg *PgStorage) SaveUser(ctx context.Context, user *models.User) (*models.User, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "pg.PgStorage.SaveUser").With("user_email", user.Email)

	stmt, err := pg.db.PreparexContext(ctx, `
    	WITH u AS (
			INSERT INTO users (name, lastname, email, country_fk, is_public, image, password, birthday)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING *
		)
		SELECT u.id, u.name, u.lastname, u.email, u.is_public, u.image, u.password, u.birthday,
			   countries.id, countries.name, countries.alpha2, countries.alpha3, countries.region
		FROM u
		JOIN countries ON countries.id = country_fk`,
	)
	if err != nil {
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	row := stmt.QueryRowxContext(ctx, user.Name, user.Lastname, user.Email, user.Country.Id, user.IsPublic, user.Image, user.PasswordHash, user.Birthday)
	if row.Err() != nil {
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	var saved models.User
	row.Scan(&saved.Id, &saved.Name, &saved.Lastname, &saved.Email, &saved.IsPublic, &saved.Image, &saved.PasswordHash, &saved.Birthday,
		&saved.Country.Id, &saved.Country.Name, &saved.Country.Alpha2, &saved.Country.Alpha3, &saved.Country.Region)

	return &saved, nil
}

func (pg *PgStorage) UserByEmail(ctx context.Context, email string) (*models.User, ero.Error) {
	return pg.userBy(ctx, "users.email = $1", email)
}

func (pg *PgStorage) UserById(ctx context.Context, id uint64) (*models.User, ero.Error) {
	return pg.userBy(ctx, "users.id = $1", id)
}

func (pg *PgStorage) userBy(ctx context.Context, where string, args ...any) (*models.User, ero.Error) {
	logCtx := erolog.NewContextBuilder().WithParent(ctx).With("op", "pg.PgStorage.userBy").With("args", args)

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

	row := stmt.QueryRowxContext(ctx, args...)
	err = row.Err()
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeNotFound, storage.ErrNoRows)
	case err != nil:
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	var user models.User
	row.Scan(&user.Id, &user.Name, &user.Lastname, &user.Email, &user.IsPublic, &user.Image, &user.PasswordHash, &user.Birthday,
		&user.Country.Id, &user.Country.Name, &user.Country.Alpha2, &user.Country.Alpha3, &user.Country.Region)

	return &user, nil
}
