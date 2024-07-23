package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func IdParam(paramName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			idStr := strings.TrimPrefix(c.Param(paramName), "id")
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				c.JSONBlob(http.StatusNotFound, ErrorMessage(fmt.Sprintf("%s is not integer", paramName)).Blob())
				return err
			}

			c.Set(paramName, id)

			return next(c)
		}
	}
}
