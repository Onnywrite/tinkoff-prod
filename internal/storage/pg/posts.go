package pg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
	"github.com/jackc/pgx/v5/pgconn"
)

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

func (pg *PgStorage) Posts(ctx context.Context, offset, count int) ([]models.Post, ero.Error) {
	logCtx := erolog.NewContextBuilder().WithParent(ctx).With("op", "pg.PgStorage.Posts").With("offset", offset).With("count", count)

	stmt, err := pg.db.PreparexContext(ctx, `
		SELECT posts.id, posts.content, posts.images_urls, posts.published_at, posts.updated_at,
			   users.id, users.name, users.lastname, users.email, users.is_public, users.image, users.password, users.birthday,
			   countries.id, countries.name, countries.alpha2, countries.alpha3, countries.region
		FROM posts
		JOIN users ON posts.author_fk = users.id
		JOIN countries ON users.country_fk = countries.id
		WHERE users.is_public = true
		ORDER BY published_at DESC
		OFFSET $1
		LIMIT $2`,
	)
	if err != nil {
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	rows, err := stmt.QueryxContext(ctx, offset, count)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeNotFound, storage.ErrNoRows)
	case err != nil:
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}
	defer rows.Close()

	posts := make([]models.Post, 0, count)
	for rows.Next() {
		var p models.Post
		err = rows.Scan(&p.Id, &p.Content, &p.ImagesUrls, &p.PublishedAt, &p.UpdatedAt,
			&p.Author.Id, &p.Author.Name, &p.Author.Lastname, &p.Author.Email, &p.Author.IsPublic,
			&p.Author.Image, &p.Author.PasswordHash, &p.Author.Birthday,
			&p.Author.Country.Id, &p.Author.Country.Name, &p.Author.Country.Alpha2,
			&p.Author.Country.Alpha3, &p.Author.Country.Region)
		if err != nil {
			return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
		}
		posts = append(posts, p)
	}
	if len(posts) == 0 {
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeNotFound, storage.ErrNoRows)
	}

	return posts, nil
}

func (pg *PgStorage) PostsNum(ctx context.Context) (uint64, ero.Error) {
	return pg.EstimatePostsNum(ctx)
}

func (pg *PgStorage) EstimatePostsNum(ctx context.Context) (uint64, ero.Error) {
	logCtx := erolog.NewContextBuilder().WithParent(ctx).With("op", "pg.PgStorage.EstimatePostsNum")

	stmt, err := pg.db.PreparexContext(ctx, `SELECT reltuples FROM pg_class where relname = 'posts'`)
	if err != nil {
		return 0, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}
	var estimate float64
	err = stmt.GetContext(ctx, &estimate)
	if err != nil {
		return 0, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	return uint64(estimate), nil
}
