package middleware

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func Pagination(defaultPageSize uint64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			page, err := strconv.ParseUint(c.QueryParam("page"), 10, 32)
			if err != nil || page < 1 {
				page = 1
			}
			pageSize, err := strconv.ParseUint(c.QueryParam("page_size"), 10, 32)
			if err != nil || pageSize < 1 {
				pageSize = defaultPageSize
			}

			c.Set("page", page)
			c.Set("page_size", pageSize)

			return next(c)
		}
	}
}
