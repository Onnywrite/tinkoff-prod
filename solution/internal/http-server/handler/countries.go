package handler

import (
	"fmt"
	"net/http"
	"regexp"

	"solution/internal/models"
	"solution/internal/storage"

	"github.com/labstack/echo/v4"
)

type CountriesProvider interface {
	Countries(regions ...string) ([]models.Country, error)
}

type CountryProvider interface {
	Country(alpha2 string) (models.Country, error)
}

func GetCountries(provider CountriesProvider) func(c echo.Context) error {
	return func(c echo.Context) error {
		regions := c.QueryParams()["region"]

		cs, err := provider.Countries(regions...)

		if err == storage.ErrCountriesNotFound {
			c.JSON(http.StatusNotFound, &crush{
				Reason: fmt.Sprintf("could not find countries within regions %v", regions),
			})
			return err
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, &crush{
				Reason: "internal error",
			})
			return err
		}

		return c.JSON(http.StatusOK, &cs)
	}
}

func GetCountriesAlpha(provider CountryProvider) func(c echo.Context) error {
	alphaRegex := regexp.MustCompile(`^[A-Z]{2}$`)

	return func(c echo.Context) error {
		alpha := c.Param("alpha2")
		if !alphaRegex.MatchString(alpha) {
			return c.JSON(http.StatusNotFound, &crush{
				Reason: "code does not seem to be an alpha2",
			})
		}

		ctr, err := provider.Country(alpha)
		if err != nil {
			c.JSON(http.StatusNotFound, &crush{
				Reason: fmt.Sprintf("could not find countries with alpha2 '%s'", alpha),
			})
			return err
		}

		return c.JSON(http.StatusOK, &ctr)
	}
}
