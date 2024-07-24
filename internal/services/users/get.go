package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

func (s *Service) UserById(ctx context.Context, id uint64, hasFullAccess bool) (PrivateOrPublicProfile, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "users.Service.UserById").With("id", id)

	user, eroErr := s.d.ByIdProvider.UserById(context.TODO(), id)
	switch {
	case errors.Is(eroErr, storage.ErrNoRows):
		s.log.DebugContext(logCtx.BuildContext(), "user not found")
		return PrivateOrPublicProfile{}, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeNotFound, ErrUserNotFound)
	case eroErr != nil:
		s.log.ErrorContext(logCtx.BuildContext(), "error while getting user by id")
		return PrivateOrPublicProfile{}, ero.New(logCtx.With("error", eroErr).Build(), ero.CodeInternal, ErrInternal)
	}

	if !user.IsPublic && !hasFullAccess {
		private := GetPrivateProfile(user)
		return PrivateOrPublicProfile{
			Private: &private,
		}, nil
	}

	public := GetProfile(user)
	return PrivateOrPublicProfile{
		Public: &public,
	}, nil
}

type PrivateOrPublicProfile struct {
	Public  *Profile
	Private *PrivateProfile
}

func (p *PrivateOrPublicProfile) Switch(ifPublic func(profile *Profile) error, ifPrivate func(profile *PrivateProfile) error) error {
	if p.Public != nil {
		return ifPublic(p.Public)
	} else if p.Private != nil {
		return ifPrivate(p.Private)
	}
	return fmt.Errorf("no profile, you shoulda handled that err")
}
