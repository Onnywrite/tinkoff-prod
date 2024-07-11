package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"

	"github.com/go-playground/validator/v10"
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

		user, err := registrator.SaveUser(context.TODO(), &models.User{
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
			pair, err := createTokens(user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, &crush{
					Reason: "error while generating tokens",
				})
				return err
			}

			c.JSON(http.StatusOK, &tokensResponse{
				Profile: getProfile(user),
				Pair:    pair,
			})
			return nil
		}

		c.JSON(status, &crush{
			Reason: err.Error(),
		})
		return err
	}
}

func createTokens(usr *models.User) (tokens.Pair, error) {
	access := tokens.Access{
		Id:    usr.Id,
		Email: usr.Email,
	}
	refresh := tokens.Refresh{
		Id: usr.Id,
	}

	accessStr, err := access.Sign()
	if err != nil {
		return tokens.Pair{}, err
	}

	refreshStr, err := refresh.Sign()
	if err != nil {
		return tokens.Pair{}, err
	}

	return tokens.Pair{
		Access:  accessStr,
		Refresh: refreshStr,
	}, nil
}
