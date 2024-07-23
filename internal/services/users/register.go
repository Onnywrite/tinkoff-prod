package users

import (
	"context"
	"errors"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/services/countries"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Register(ctx context.Context, userData RegisterData) (*AuthorizedUser, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "users.Service.Register").WithSecret("email", userData.Email, 50)

	if err := userData.Validate(); err != nil {
		s.log.DebugContext(err.Context(ctx), "errors validating user data")
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.ErrorContext(logCtx.With("error", err).BuildContext(), "error while hashing password")
		return nil, ero.New(logCtx.Build(), ero.CodeInternal, ErrInternal)
	}

	user, eroErr := s.d.Saver.SaveUser(ctx, &models.User{
		Name:     userData.Name,
		Lastname: userData.Lastname,
		Email:    userData.Email,
		Country: models.Country{
			Id: userData.CountryId,
		},
		IsPublic:     *userData.IsPublic,
		Image:        userData.Image,
		PasswordHash: string(hash),
		Birthday:     time.Time(userData.Birthday),
	})
	switch {
	case errors.Is(eroErr, storage.ErrUniqueConstraint):
		s.log.DebugContext(logCtx.BuildContext(), "user already exists")
		return nil, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeExists, ErrUserExists)
	case errors.Is(eroErr, storage.ErrForeignKeyConstraint):
		s.log.DebugContext(logCtx.BuildContext(), "country with given id does not exist")
		return nil, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeNotFound, countries.ErrCountryNotFound)
	case eroErr != nil:
		s.log.ErrorContext(eroErr.Context(ctx), "internal error")
		return nil, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeInternal, ErrInternal)
	}

	pair, err := tokens.NewPair(user, 0)
	if err != nil {
		s.log.ErrorContext(logCtx.With("error", err).BuildContext(), "error while generating tokens")
		return nil, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeInternal, ErrInternal)
	}

	return &AuthorizedUser{
		Profile: GetProfile(user),
		Pair:    pair,
	}, nil
}
