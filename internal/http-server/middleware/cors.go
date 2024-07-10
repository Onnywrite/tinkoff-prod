package middleware

import (
	"github.com/labstack/echo/v4"
)

func Cors() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Add("Access-Control-Allow-Origin", "*")
			c.Response().Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			c.Response().Header().Add("Access-Control-Allow-Headers", "Content-Type")
			c.Response().Header().Add("Access-Control-Max-Age", "86400")

			return next(c)
		}
	}
}
