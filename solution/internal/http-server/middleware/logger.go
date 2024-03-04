package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func Logger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			const op = "middleware.Logger"
			log := logger.With(slog.String("op", op))

			body, err := io.ReadAll(c.Request().Body)
			if err != nil {
				body = []byte{}
			}

			t := time.Now()

			errText := ""
			if err = next(c); err != nil {
				c.Error(err)
				errText = err.Error()
			}
			end := int(time.Since(t) / time.Millisecond)

			log.Info("request",
				slog.String("uri", c.Request().RequestURI),
				slog.String("method", c.Request().Method),
				slog.String("body", string(body)),
				slog.Int("code", c.Response().Status),
				slog.String("status", http.StatusText(c.Response().Status)),
				slog.Int("elapsed_ms", end),
				slog.String("content_type", c.Response().Header()["Content-Type"][0]),
				slog.String("error", errText),
			)

			return nil
		}
	}
}
