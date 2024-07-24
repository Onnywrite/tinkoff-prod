package users

import (
	"context"
	"errors"

	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

func (s *Service) Refresh(ctx context.Context, refresh tokens.RefreshString) (*AuthorizedUser, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "users.Service.Refresh")

	token, err := refresh.ParseVerify()
	switch {
	case errors.Is(err, tokens.ErrExpired):
		s.log.DebugContext(logCtx.BuildContext(), "refresh token has expired")
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeUnauthorized, err)
	case err != nil:
		s.log.ErrorContext(logCtx.BuildContext(), "could not parse refresh token")
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeUnauthorized, ErrInvalidToken)
	}

	user, eroErr := s.d.ByIdProvider.UserById(ctx, token.Id)
	switch {
	case errors.Is(eroErr, storage.ErrNoRows):
		s.log.DebugContext(logCtx.BuildContext(), "user not found")
		return nil, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeNotFound, ErrUserNotFound)
	case eroErr != nil:
		s.log.ErrorContext(logCtx.BuildContext(), "internal error")
		return nil, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeInternal, ErrInternal)
	}

	pair, err := tokens.NewPair(user, 0)
	if err != nil {
		s.log.ErrorContext(logCtx.BuildContext(), "error while generating tokens")
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, ErrInternal)
	}

	return &AuthorizedUser{
		Profile: GetProfile(user),
		Pair:    pair,
	}, nil
}
