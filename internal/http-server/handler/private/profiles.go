package privatehandler

import (
	"context"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/services/users"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/labstack/echo/v4"
)

func GetProfile(provider UserProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId := c.Get("user_id").(uint64)
		id := c.Get("id").(uint64)
		privateOrPublic, err := provider.UserById(context.TODO(), userId, userId == id)
		if err != nil {
			return c.JSONBlob(ero.ToHttpCode(err.Code()), []byte(err.Error()))
		}

		return privateOrPublic.Switch(
			func(profile *users.Profile) error {
				return c.JSON(http.StatusOK, profile)
			},
			func(profile *users.PrivateProfile) error {
				return c.JSON(http.StatusOK, profile)
			},
		)
	}
}
