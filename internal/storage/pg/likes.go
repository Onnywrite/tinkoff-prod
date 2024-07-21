package pg

import (
	"context"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

func (pg *PgStorage) SaveLike(ctx context.Context, like models.Like) ero.Error {
	logCtx := erolog.NewContextBuilder().With("op", "pg.PgStorage.SaveLike").With("user_id", like.User.Id).With("post_id", like.Post.Id)

	stmt, err := pg.db.PreparexContext(ctx, `
		INSERT INTO likes (user_fk, post_fk)
		VALUES ($1, $2)`,
	)
	if err != nil {
		return ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	_, err = stmt.ExecContext(ctx, like.User.Id, like.Post.Id)
	if err != nil {
		return ero.New(logCtx.With("error", err).Build(), ero.CodeUnknownServer, getError(err))
	}

	return nil
}

func (pg *PgStorage) DeleteLike(ctx context.Context, like models.Like) ero.Error {
	logCtx := erolog.NewContextBuilder().With("op", "pg.PgStorage.SaveLike").With("user_id", like.User.Id).With("post_id", like.Post.Id)

	stmt, err := pg.db.PreparexContext(ctx, `
		DELETE FROM likes
		WHERE user_fk = $1 AND post_fk = $2`,
	)
	if err != nil {
		return ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
	}

	res, err := stmt.ExecContext(ctx, like.User.Id, like.Post.Id)
	if err != nil {
		return ero.New(logCtx.With("error", err).Build(), ero.CodeUnknownServer, getError(err))
	}

	if res != nil {
		if affected, _ := res.RowsAffected(); affected == 0 {
			return ero.New(logCtx.With("error", err).Build(), ero.CodeNotFound, storage.ErrNoRows)
		}
	}

	return nil
}

func (pg *PgStorage) Likes(ctx context.Context, offset, count int, postId uint64) (<-chan models.Like, <-chan ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "pg.PgStorage.SaveLike").With("offset", offset).With("count", count).With("post_id", postId)

	likesCh := make(chan models.Like, 10)
	errChan := make(chan ero.Error, 1)

	go func() {
		defer close(likesCh)
		defer close(errChan)

		stmt, err := pg.db.PreparexContext(ctx, `
			SELECT users.id, users.name, users.lastname, users.email, country_fk,
				   users.is_public, users.image, users.password, users.birthday, liked_at
			FROM likes
			JOIN users ON users.id = user_fk
			WHERE post_fk = $3
			ORDER BY liked_at DESC
			OFFSET $1
			LIMIT $2`,
		)
		if err != nil {
			errChan <- ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
			return
		}

		rows, err := stmt.QueryxContext(ctx, offset, count, postId)
		if err != nil {
			errChan <- ero.New(logCtx.With("error", err).Build(), ero.CodeUnknownServer, getError(err))
			return
		}
		defer rows.Close()

		for rows.Next() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			var like models.Like
			err = rows.Scan(&like.User.Id, &like.User.Name, &like.User.Lastname, &like.User.Email, &like.User.Country.Id,
				&like.User.IsPublic, &like.User.Image, &like.User.PasswordHash, &like.User.Birthday, &like.LikedAt)
			like.Post.Id = postId
			if err != nil {
				errChan <- ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, storage.ErrInternal)
				return
			}

			select {
			case likesCh <- like:
			case <-ctx.Done():
				return
			}
		}
	}()

	return likesCh, errChan
}

func (pg *PgStorage) LikesNum(ctx context.Context, postId uint64) (uint64, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "pg.PgStorage.LikesNum").With("post_id", postId)

	var count uint64
	err := pg.db.GetContext(ctx, &count, `
		SELECT COUNT(*)
		FROM likes
		WHERE post_fk = $1`,
		postId,
	)
	if err != nil {
		return 0, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, getError(err))
	}

	return count, nil
}
