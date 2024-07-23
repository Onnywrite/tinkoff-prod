package privatehandler

import (
	"context"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/services/users"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/labstack/echo/v4"
)

type UserProvider interface {
	UserById(ctx context.Context, id uint64, hasFullAccess bool) (users.PrivateOrPublicProfile, ero.Error)
}

func GetMe(provider UserProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		privateOrPublic, err := provider.UserById(context.TODO(), c.Get("id").(uint64), true)
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
