package handler

import (
	"net/http"

	"solution/internal/models"

	"github.com/labstack/echo/v4"
)

type ProfileProvider interface {
	Profile(login string) (*models.Profile, error)
}

func GetMeProfile(provider ProfileProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		profile, err := provider.Profile(c.Get("login").(string))
		if err != nil {
			c.JSON(http.StatusNotFound, &crush{
				Reason: "profile not found",
			})
			return err
		}

		return c.JSON(http.StatusOK, profile)
	}
}
