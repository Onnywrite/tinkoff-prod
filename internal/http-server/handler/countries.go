package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"

	"github.com/labstack/echo/v4"
)

type CountriesProvider interface {
	Countries(ctx context.Context, regions ...string) ([]models.Country, ero.Error)
}

type CountryProvider interface {
	Country(ctx context.Context, alpha2 string) (models.Country, ero.Error)
}

func capitalizeFirstLetter(s string) string {
	runes := []rune(strings.ToLower(strings.Trim(s, "\" ")))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func GetCountries(provider CountriesProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		regions := c.QueryParams()["region"]
		for i := range regions {
			regions[i] = capitalizeFirstLetter(regions[i])
		}

		cs, err := provider.Countries(context.TODO(), regions...)
		switch {
		case errors.Is(err, storage.ErrNoRows):
			c.JSONBlob(http.StatusNotFound, errorMessage(fmt.Sprintf("could not find countries within regions %v", regions)).Blob())
			return err
		case err != nil:
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
			return err
		}

		return c.JSON(http.StatusOK, &cs)
	}
}

func GetCountryAlpha(provider CountryProvider) echo.HandlerFunc {
	alphaRegex := regexp.MustCompile(`^[A-Za-z]{2}$`)

	return func(c echo.Context) error {
		alpha := c.Param("alpha2")
		if !alphaRegex.MatchString(alpha) {
			return c.JSONBlob(http.StatusNotFound, errorMessage("code does not seem to be an alpha2").Blob())
		}
		alpha = strings.ToUpper(alpha)

		ctr, err := provider.Country(context.TODO(), alpha)
		switch {
		case errors.Is(err, storage.ErrNoRows):
			c.JSONBlob(http.StatusNotFound, errorMessage(fmt.Sprintf("could not find countries with alpha2 '%s'", alpha)).Blob())
			return err
		case err != nil:
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
			return err
		}

		return c.JSON(http.StatusOK, &ctr)
	}
}
