package countries

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"unicode"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

type Service struct {
	log       *slog.Logger
	sprovider CountriesProvider
	provider  CountryProvider
}

type CountriesProvider interface {
	Countries(ctx context.Context, regions ...string) ([]models.Country, ero.Error)
}

type CountryProvider interface {
	Country(ctx context.Context, alpha2 string) (models.Country, ero.Error)
}

func New(logger *slog.Logger, provider CountriesProvider, cProvider CountryProvider) *Service {
	return &Service{
		log:       logger,
		sprovider: provider,
		provider:  cProvider,
	}
}

func capitalizeFirstLetter(s string) string {
	runes := []rune(strings.ToLower(strings.Trim(s, "\" ")))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func (s *Service) Countries(ctx context.Context, regions ...string) ([]models.Country, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "countries.Service.Countries")

	for i := range regions {
		regions[i] = capitalizeFirstLetter(regions[i])
	}

	cs, err := s.sprovider.Countries(ctx, regions...)
	switch {
	case errors.Is(err, storage.ErrNoRows):
		s.log.DebugContext(logCtx.BuildContext(), "could not find countries within given regions")
		return nil, ero.New(logCtx.WithParent(err.Context(ctx)).Build(), ero.CodeNotFound, ErrCountriesNotFound)
	case err != nil:
		s.log.ErrorContext(err.Context(ctx), "could not get countries")
		return nil, ero.New(logCtx.WithParent(err.Context(ctx)).With("error", err).Build(), ero.CodeInternal, ErrInternal)
	}

	return cs, nil
}

var alphaRegex = regexp.MustCompile(`^[A-Z]{2}$`)

func (s *Service) Country(ctx context.Context, alpha2 string) (models.Country, ero.Error) {
	logCtx := erolog.NewContextBuilder().With("op", "countries.Service.Country").With("alpha2", alpha2)

	alpha2 = strings.ToUpper(alpha2)
	if !alphaRegex.MatchString(alpha2) {
		s.log.DebugContext(logCtx.BuildContext(), "alpha2 does not seem to be an alpha2")
		return models.Country{}, ero.New(logCtx.Build(), ero.CodeBadRequest, ErrBadAlpha2)
	}

	ctr, err := s.provider.Country(context.TODO(), alpha2)
	switch {
	case errors.Is(err, storage.ErrNoRows):
		s.log.DebugContext(logCtx.BuildContext(), "country with given alpha2 does not exist")
		return models.Country{}, ero.New(logCtx.Build(), ero.CodeNotFound, ErrCountryNotFound)
	case err != nil:
		s.log.ErrorContext(err.Context(ctx), "could not get country")
		return models.Country{}, ero.New(logCtx.WithParent(err.Context(ctx)).With("error", err).Build(), ero.CodeInternal, ErrInternal)
	}

	return ctr, nil
}
