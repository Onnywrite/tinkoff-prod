package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/services/countries"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"

	"github.com/labstack/echo/v4"
)

type CountriesProvider interface {
	Countries(ctx context.Context, regions ...string) ([]models.Country, ero.Error)
}

type CountryProvider interface {
	Country(ctx context.Context, alpha2 string) (models.Country, ero.Error)
}

func GetCountries(provider CountriesProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		regions := c.QueryParams()["region"]

		cs, err := provider.Countries(context.TODO(), regions...)
		switch {
		case errors.Is(err, countries.ErrCountriesNotFound):
			c.JSONBlob(http.StatusNotFound, ErrorMessage(fmt.Sprintf("could not find countries within regions %v", regions)).Blob())
			return err
		case err != nil:
			c.JSONBlob(http.StatusInternalServerError, ErrorMessage("internal error").Blob())
			return err
		}

		return c.JSON(http.StatusOK, &cs)
	}
}

func GetCountryAlpha(provider CountryProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		alpha := c.Param("alpha2")

		ctr, err := provider.Country(context.TODO(), alpha)
		switch {
		case errors.Is(err, countries.ErrCountryNotFound):
			c.JSONBlob(http.StatusNotFound, ErrorMessage(fmt.Sprintf("could not find countries with alpha2 '%s'", alpha)).Blob())
			return err
		case errors.Is(err, countries.ErrBadAlpha2):
			c.JSONBlob(http.StatusBadRequest, ErrorMessage("alpha2 is not valid").Blob())
			return err
		case err != nil:
			c.JSONBlob(http.StatusInternalServerError, ErrorMessage("internal error").Blob())
			return err
		}

		return c.JSON(http.StatusOK, &ctr)
	}
}
