package authhandler

import (
	"context"
	"net/http"

	"github.com/Onnywrite/tinkoff-prod/internal/http-server/handler"
	"github.com/Onnywrite/tinkoff-prod/internal/services/users"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"

	"github.com/labstack/echo/v4"
)

type UserRegistrator interface {
	Register(ctx context.Context, userData users.RegisterData) (*users.AuthorizedUser, ero.Error)
}

func PostRegister(registrator UserRegistrator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var u users.RegisterData
		if err := c.Bind(&u); err != nil {
			c.JSONBlob(http.StatusBadRequest, handler.ErrorMessage("could not bind the body").Blob())
			return err
		}

		authUser, eroErr := registrator.Register(context.TODO(), u)
		if eroErr != nil {
			c.JSONBlob(ero.ToHttpCode(eroErr.Code()), []byte(eroErr.Error()))
			return eroErr
		}

		c.JSON(http.StatusOK, authUser)
		return nil
	}
}
