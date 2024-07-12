package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
	"github.com/labstack/echo/v4"
)

func Logger(logger *slog.Logger) echo.MiddlewareFunc {
	const op = "middleware.Logger"
	log := logger.With(slog.String("op", op))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			t := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}
			end := int(time.Since(t) / time.Millisecond)

			level := statusToLevel(c.Response().Status)

			ctx := context.Background()
			if eroErr, ok := err.(ero.Error); ok {
				ctx = eroErr.Context(ctx)
			} else if err != nil {
				// TODO: refactor - always return ero.Error
				ctx = erolog.NewContextBuilder().With("error", err.Error()).BuildContext()
			}

			log.LogAttrs(ctx, level, "request",
				slog.String("uri", c.Request().RequestURI),
				slog.String("method", c.Request().Method),
				slog.Int("code", c.Response().Status),
				slog.String("status", http.StatusText(c.Response().Status)),
				slog.Int("elapsed_ms", end),
				slog.String("content_type", c.Response().Header()["Content-Type"][0]),
				// slog.String("error", err.Error()),
			)

			return nil
		}
	}
}

func statusToLevel(status int) slog.Level {
	level := slog.LevelInfo
	if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
		level = slog.LevelWarn
	}
	if status >= http.StatusInternalServerError {
		level = slog.LevelError
	}
	return level
}
