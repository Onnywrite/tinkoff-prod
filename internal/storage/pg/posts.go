package pg

import (
	"context"
	"fmt"
	"slices"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
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

	if err != nil {
		return 0, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, getError(err))
	}

	return id, nil
}

func (pg *PgStorage) Posts(ctx context.Context, offset, count int) (<-chan models.Post, <-chan ero.Error) {
	return pg.postsBy(ctx, offset, count, "users.is_public = true")
}

func (pg *PgStorage) UsersPosts(ctx context.Context, offset, count int, userId uint64) (<-chan models.Post, <-chan ero.Error) {
	return pg.postsBy(ctx, offset, count, "posts.author_fk = $3", userId)
}

func (pg *PgStorage) postsBy(ctx context.Context, offset, count int, where string, args ...any) (<-chan models.Post, <-chan ero.Error) {
	logCtx := erolog.NewContextBuilder().WithParent(ctx).With("op", "pg.PgStorage.Posts").With("offset", offset).With("count", count)

	posts := make(chan models.Post, 10)
	errChan := make(chan ero.Error, 1)

	go func() {
		defer close(posts)
		defer close(errChan)

		stmt, err := pg.db.PreparexContext(ctx, fmt.Sprintf(`
			SELECT posts.id, posts.content, posts.images_urls, posts.published_at, posts.updated_at,
				   users.id, users.name, users.lastname, users.email, users.is_public, users.image, users.password, users.birthday,
				   countries.id, countries.name, countries.alpha2, countries.alpha3, countries.region
			FROM posts
			JOIN users ON posts.author_fk = users.id
			JOIN countries ON users.country_fk = countries.id
			WHERE %s
			ORDER BY published_at DESC
			OFFSET $1
			LIMIT $2`, where),
		)
		if err != nil {
			errChan <- ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
			return
		}

		rows, err := stmt.QueryxContext(ctx, slices.Concat([]any{offset, count}, args)...)
		if err != nil {
			errChan <- ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
			return
		}
		defer rows.Close()

		for rows.Next() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			var p models.Post
			err = rows.Scan(&p.Id, &p.Content, &p.ImagesUrls, &p.PublishedAt, &p.UpdatedAt,
				&p.Author.Id, &p.Author.Name, &p.Author.Lastname, &p.Author.Email, &p.Author.IsPublic,
				&p.Author.Image, &p.Author.PasswordHash, &p.Author.Birthday,
				&p.Author.Country.Id, &p.Author.Country.Name, &p.Author.Country.Alpha2,
				&p.Author.Country.Alpha3, &p.Author.Country.Region)
			if err != nil {
				errChan <- ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
				return
			}

			select {
			case posts <- p:
			case <-ctx.Done():
				return
			}
		}
	}()

	return posts, errChan
}

func (pg *PgStorage) UsersPostsNum(ctx context.Context, userId uint64) (uint64, ero.Error) {
	return pg.postsNum(ctx, "WHERE posts.author_fk = $1", userId)
}

func (pg *PgStorage) PostsNum(ctx context.Context) (uint64, ero.Error) {
	return pg.postsNum(ctx, "JOIN users ON author_fk = users.id WHERE users.is_public = true")
}

func (pg *PgStorage) postsNum(ctx context.Context, sql string, args ...any) (uint64, ero.Error) {
	logCtx := erolog.NewContextBuilder().WithParent(ctx).With("op", "pg.PgStorage.postsNum")

	stmt, err := pg.db.PreparexContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM posts %s`, sql))
	if err != nil {
		return 0, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}
	var estimate float64
	err = stmt.GetContext(ctx, &estimate, args...)
	if err != nil {
		return 0, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	return uint64(estimate), nil
}
