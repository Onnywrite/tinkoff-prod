package authhandler

import (
	"context"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/http-server/handler"
	"github.com/Onnywrite/tinkoff-prod/internal/services/users"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/labstack/echo/v4"
)

type IdentityProvider interface {
	SignIn(ctx context.Context, creds users.Credentials) (*users.AuthorizedUser, ero.Error)
}

func PostSignIn(provider IdentityProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		var data users.Credentials
		if err := c.Bind(&data); err != nil {
			c.JSONBlob(http.StatusBadRequest, handler.ErrorMessage("could not bind the body").Blob())
			return err
		}

		authUser, eroErr := provider.SignIn(context.TODO(), data)
		if eroErr != nil {
			c.JSONBlob(ero.ToHttpCode(eroErr.Code()), []byte(eroErr.Error()))
			return eroErr
		}

		c.JSON(http.StatusOK, authUser)

		return nil
	}
}
