package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserRegistrator interface {
	SaveUser(ctx context.Context, user *models.User) (*models.User, error)
}

type dateOnly time.Time

func (d *dateOnly) UnmarshalJSON(b []byte) error {
	ss := strings.Split(strings.Trim(string(b), "\""), "-")
	if len(ss) != 3 {
		return fmt.Errorf("invalid date")
	}

	yyyy, err := strconv.ParseInt(ss[0], 10, 32)
	if err != nil {
		return err
	}
	mm, err := strconv.ParseInt(ss[1], 10, 32)
	if err != nil {
		return err
	}
	dd, err := strconv.ParseInt(ss[2], 10, 32)
	if err != nil {
		return err
	}
	*d = dateOnly(time.Date(int(yyyy), time.Month(mm), int(dd), 0, 0, 0, 0, time.UTC))

	return nil
}

func PostRegister(registrator UserRegistrator) echo.HandlerFunc {
	type registerData struct {
		Name     string   `validate:"required,min=2,max=32" json:"name"`
		Lastname string   `validate:"required,min=2,max=32" json:"surname"`
		Email    string   `validate:"required,email" json:"email"`
		Password string   `validate:"required,min=8" json:"password"`
		Birthday dateOnly `json:"birthday"`
	}
	return func(c echo.Context) error {
		var u registerData
		if err := c.Bind(&u); err != nil {
			c.JSON(http.StatusInternalServerError, &crush{
				Reason: "could not bind the body",
			})
			return err
		}

		validate := validator.New()

		err := validate.StructCtx(context.TODO(), &u)
		if err != nil {
			if ve, ok := err.(validator.ValidationErrors); ok {
				c.JSON(http.StatusBadRequest, echo.Map{
					"reason": "validation errors",
					"fields": ve.Error(),
				})
				return err
			}
			c.JSON(http.StatusInternalServerError, &crush{
				Reason: "validation error",
			})
			return err
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		profile, err := registrator.SaveUser(context.TODO(), &models.User{
			Name:     u.Name,
			Lastname: u.Lastname,
			Email:    u.Email,
			Country: models.Country{
				Id: 70,
			},
			IsPublic:     true,
			Image:        "",
			PasswordHash: string(hash),
			Birthday:     time.Time(u.Birthday),
		})
		status := http.StatusOK
		switch {
		case err == storage.ErrInternal:
			status = http.StatusInternalServerError
		case err != nil:
			status = http.StatusConflict
		default:
			access, refresh, err := createTokens(profile, []byte("$my_%SUPER(n0t-so=MUch)_secret123"))
			if err != nil {
				c.JSON(http.StatusInternalServerError, &crush{
					Reason: "error while generating tokens",
				})
				return err
			}

			c.JSON(http.StatusOK, echo.Map{
				"refresh": refresh,
				"access":  access,
			})
			return nil
		}

		c.JSON(status, &crush{
			Reason: err.Error(),
		})
		return err
	}
}

type UserProvider interface {
	UserByEmail(ctx context.Context, email string) (*models.User, error)
}

func PostSignIn(provider UserProvider) echo.HandlerFunc {
	type loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(c echo.Context) error {
		var data loginData
		if err := c.Bind(&data); err != nil {
			c.JSON(http.StatusUnauthorized, &crush{
				Reason: "could not bind the body",
			})
			return err
		}

		usr, err := provider.UserByEmail(context.TODO(), data.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, &crush{
				Reason: "invalid email or password",
			})
			return err
		}

		if err = bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(data.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, &crush{
				Reason: "invalid email or password",
			})
			return err
		}

		access, refresh, err := createTokens(usr, []byte("$my_%SUPER(n0t-so=MUch)_secret123"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, &crush{
				Reason: "error while generating tokens",
			})
			return err
		}

		c.JSON(http.StatusOK, echo.Map{
			"refresh": refresh,
			"access":  access,
		})

		return nil
	}
}

func createTokens(usr *models.User, secret []byte) (access string, refresh string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    usr.Id,
		"email": usr.Email,
		// TODO: exp config
		"exp": time.Now().Add(5 * time.Minute).Unix(),
	})

	// TODO: secret config
	access, err = token.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": usr.Id,
		// TODO: exp config
		"exp": time.Now().Add(168 * time.Hour).Unix(),
	})

	refresh, err = token.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	return
}
