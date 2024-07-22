package users

import (
	"context"
	"errors"

	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) SignIn(ctx context.Context, creds Credentials) (*AuthorizedUser, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "users.Service.SignIn").WithSecret("email", creds.Email, 50)

	user, eroErr := s.d.ByEmailProvider.UserByEmail(ctx, creds.Email)
	switch {
	case errors.Is(eroErr, storage.ErrNoRows):
		s.log.DebugContext(logCtx.BuildContext(), "invalid email or password")
		return nil, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeUnauthorized, ErrInvalidCredentials)
	case eroErr != nil:
		s.log.ErrorContext(logCtx.BuildContext(), "internal error")
		return nil, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeInternal, ErrInternal)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		s.log.ErrorContext(logCtx.With("error", err).BuildContext(), "invalid email or password")
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeUnauthorized, ErrInvalidCredentials)
	}

	pair, err := tokens.NewPair(user, 0)
	if err != nil {
		s.log.ErrorContext(logCtx.With("error", err).BuildContext(), "error while generating tokens")
		return nil, ero.New(logCtx.With("error", err).Build(), ero.CodeInternal, ErrInternal)
	}

	return &AuthorizedUser{
		Profile: GetProfile(user),
		Pair:    pair,
	}, nil
}
