package handler

import (
	"net/http"

	"solution/internal/models"
	"solution/internal/storage"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserRegistrator interface {
	RegisterUser(user *models.User) (*models.Profile, error)
}

func PostRegister(registrator UserRegistrator) echo.HandlerFunc {
	validateList := [][2]string{
		{"Login", "invalid login, cannot be 'me'"},
		{"Email", "invalid email"},
		{"CountryCode", "invalid country code, length must be 2"},
		{"Phone", "invalid phone format"},
		{"Image", "image URI is too long"},
		{"Password", "password is too short, must be at least 8 symbols"},
	}

	return func(c echo.Context) error {
		var u models.User
		if err := c.Bind(&u); err != nil {
			c.JSON(http.StatusInternalServerError, &crush{
				Reason: "could not bind the body",
			})
			return err
		}

		validate := validator.New()

		validateField := func(field, failMsg string) error {
			err := validate.StructPartial(&u, field)
			c.JSON(http.StatusBadRequest, &crush{
				Reason: failMsg,
			})
			return err
		}

		for _, valid := range validateList {
			if err := validateField(valid[0], valid[1]); err != nil {
				return err
			}
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hash)

		profile, err := registrator.RegisterUser(&u)
		status := http.StatusOK
		switch {
		case err == storage.ErrInternal:
			status = http.StatusInternalServerError
		case err != nil:
			status = http.StatusConflict
		default:
			c.JSON(status, &profile)
			return nil
		}

		c.JSON(status, &crush{
			Reason: err.Error(),
		})
		return err
	}
}
