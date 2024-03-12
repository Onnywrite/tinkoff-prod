package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetPing() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}
}
