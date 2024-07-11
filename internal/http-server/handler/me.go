package handler

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetMeProfile(provider UserProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		u, err := provider.UserByEmail(context.TODO(), c.Get("email").(string))
		if err != nil {
			c.JSON(http.StatusNotFound, &crush{
				Reason: "user not found",
			})
			return err
		}

		profile := getProfile(u)

		return c.JSON(http.StatusOK, profile)
	}
}
