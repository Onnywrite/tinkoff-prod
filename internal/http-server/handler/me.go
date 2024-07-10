package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/models"

	"github.com/labstack/echo/v4"
)

func GetMeProfile(provider UserProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		u, err := provider.UserByEmail(context.TODO(), c.Get("email").(string))
		if err != nil {
			c.JSON(http.StatusNotFound, &crush{
				Reason: "user not found",
			})
			return err
		}

		view := struct {
			Id       uint64         `json:"id"`
			Name     string         `json:"name"`
			Lastname string         `json:"lastname"`
			Email    string         `json:"email"`
			Country  models.Country `json:"country"`
			IsPublic bool           `json:"is_public"`
			Image    string         `json:"image"`
			Birthday string         `json:"birthday"`
		}{
			Id:       u.Id,
			Name:     u.Name,
			Lastname: u.Lastname,
			Email:    u.Email,
			Country:  u.Country,
			IsPublic: u.IsPublic,
			Image:    u.Image,
			Birthday: u.Birthday.Format(time.DateOnly),
		}
		return c.JSON(http.StatusOK, &view)
	}
}
