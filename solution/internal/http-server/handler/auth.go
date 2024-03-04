package handler

import (
	"fmt"
	"net/http"
	"time"

	"solution/internal/models"
	"solution/internal/storage"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserRegistrator interface {
	RegisterUser(user *models.User) (*models.Profile, error)
}

func PostRegister(registrator UserRegistrator) echo.HandlerFunc {
	validateList := [][2]string{
		{"Login", "invalid login, cannot be 'me' or empty"},
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
			if err != nil {
				c.JSON(http.StatusBadRequest, &crush{
					Reason: failMsg,
				})
			}
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

type UserProvider interface {
	User(login string) (*models.User, error)
}

func PostSignIn(provider UserProvider) echo.HandlerFunc {
	type loginData struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(c echo.Context) error {
		var data loginData
		fmt.Println("a")
		if err := c.Bind(&data); err != nil {
			c.JSON(http.StatusUnauthorized, &crush{
				Reason: "could not bind the body",
			})
			return err
		}

		fmt.Println("b")
		usr, err := provider.User(data.Login)
		if err != nil {
			c.JSON(http.StatusUnauthorized, &crush{
				Reason: "invalid login or password",
			})
			return err
		}

		fmt.Println("c")
		if err = bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(data.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, &crush{
				Reason: "invalid login or password",
			})
			return err
		}

		fmt.Println("d")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"login": usr.Login,
			"email": usr.Email,
			"phone": usr.Phone,
			"exp":   time.Now().Add(time.Hour).Unix(),
		})

		fmt.Println("e")
		tokenString, err := token.SignedString([]byte("$my_%SUPER(n0t-so=MUch)_secret123"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, &crush{
				Reason: "failed creating token",
			})
			return err
		}

		fmt.Println("f")
		c.JSON(http.StatusOK, echo.Map{
			"token": tokenString,
		})

		return nil
	}
}
