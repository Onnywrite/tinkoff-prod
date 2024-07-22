package pg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
	"github.com/jmoiron/sqlx"
)

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
