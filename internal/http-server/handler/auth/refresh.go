package authhandler

import (
	"context"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/http-server/handler"
	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/services/users"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/labstack/echo/v4"
)

type AccessTokenUpdater interface {
	Refresh(ctx context.Context, refresh tokens.RefreshString) (*users.AuthorizedUser, ero.Error)
}

func PostRefresh(updater AccessTokenUpdater) echo.HandlerFunc {
	type refreshToken struct {
		Refresh tokens.RefreshString `json:"refresh"`
	}

	return func(c echo.Context) error {
		var token refreshToken
		if err := c.Bind(&token); err != nil {
			c.JSONBlob(http.StatusBadRequest, handler.ErrorMessage("could not bind the body").Blob())
			return err
		}

		authUser, eroErr := updater.Refresh(context.TODO(), token.Refresh)
		if eroErr != nil {
			c.JSONBlob(ero.ToHttpCode(eroErr.Code()), []byte(eroErr.Error()))
			return eroErr
		}

		return c.JSON(http.StatusOK, authUser)
	}
}
