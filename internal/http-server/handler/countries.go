package handler

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"

	"github.com/labstack/echo/v4"
)

type CountriesProvider interface {
	Countries(ctx context.Context, regions ...string) ([]models.Country, error)
}

type CountryProvider interface {
	Country(ctx context.Context, alpha2 string) (models.Country, error)
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

func GetCountryAlpha(provider CountryProvider) echo.HandlerFunc {
	alphaRegex := regexp.MustCompile(`^[A-Za-z]{2}$`)

	return func(c echo.Context) error {
		alpha := c.Param("alpha2")
		if !alphaRegex.MatchString(alpha) {
			return c.JSON(http.StatusNotFound, &crush{
				Reason: "code does not seem to be an alpha2",
			})
		}
		alpha = strings.ToUpper(alpha)

		ctr, err := provider.Country(context.TODO(), alpha)
		if err != nil {
			c.JSON(http.StatusNotFound, &crush{
				Reason: fmt.Sprintf("could not find countries with alpha2 '%s'", alpha),
			})
			return err
		}

		return c.JSON(http.StatusOK, &ctr)
	}
}
