package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/internal/storage"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserRegistrator interface {
	SaveUser(ctx context.Context, user *models.User) (*models.User, ero.Error)
}

func PostRegister(registrator UserRegistrator) echo.HandlerFunc {
	return func(c echo.Context) error {
		var u registerData
		if err := c.Bind(&u); err != nil {
			c.JSONBlob(http.StatusBadRequest, errorMessage("could not bind the body").Blob())
			return err
		}

		eroErr := u.Validate()
		if eroErr != nil {
			c.JSON(http.StatusBadRequest, eroErr)
			return eroErr
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
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
		switch {
		case errors.Is(err, storage.ErrUniqueConstraint):
			c.JSONBlob(http.StatusConflict, errorMessage("user already exists").Blob())
			return err
		case errors.Is(err, storage.ErrForeignKeyConstraint):
			c.JSONBlob(http.StatusConflict, errorMessage(fmt.Sprintf("country with id %d does not exist", user.Country.Id)).Blob())
			return err
		case err != nil:
			c.JSONBlob(http.StatusInternalServerError, errorMessage("internal error").Blob())
			return err
		}

		pair, err := tokens.NewPair(user)
		if err != nil {
			c.JSONBlob(http.StatusInternalServerError, errorMessage("error while generating tokens").Blob())
			return err
		}

		c.JSON(http.StatusOK, &tokensResponse{
			Profile: getProfile(user),
			Pair:    pair,
		})
		return nil
	}
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

type registerData struct {
	Name     string   `json:"name"`
	Lastname string   `json:"surname"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Birthday dateOnly `json:"birthday"`
}

func (d *registerData) Validate() ero.Error {
	nameRegex := regexp.MustCompile(`^[\p{L}]+(-[\p{L}]+)*$`)
	emailRegex := regexp.MustCompile(`^[a-z0-9._-]+@[a-z0-9.-]+\.[a-z]{2,4}$`)

	type fieldError struct {
		Field    string   `json:"field"`
		Messages []string `json:"messages"`
	}

	formatName := func(name string) string {
		runes := []rune(strings.ToLower(name))
		runes[0] = unicode.ToUpper(runes[0])
		for i := 1; i < len(runes); i++ {
			if runes[i-1] == '-' {
				runes[i] = unicode.ToUpper(runes[i])
			}
		}
		return string(runes)
	}

	errorsMap := make(map[string][]string)

	errorsMap["name"] = make([]string, 0, 2)
	errorsMap["surname"] = make([]string, 0, 2)
	errorsMap["email"] = make([]string, 0, 2)
	errorsMap["password"] = make([]string, 0, 2)
	errorsMap["birthday"] = make([]string, 0, 2)

	if utf8.RuneCountInString(d.Name) > 32 {
		errorsMap["name"] = append(errorsMap["name"], "too long, must be less than 32 characters")
	}
	if !nameRegex.MatchString(d.Name) {
		errorsMap["name"] = append(errorsMap["name"], "invalid characters set")
	}
	d.Name = formatName(d.Name)

	if utf8.RuneCountInString(d.Lastname) > 32 {
		errorsMap["surname"] = append(errorsMap["surname"], "too long, must be less than 32 characters")
	}
	if !nameRegex.MatchString(d.Lastname) {
		errorsMap["surname"] = append(errorsMap["surname"], "invalid characters set")
	}
	d.Lastname = formatName(d.Lastname)

	if utf8.RuneCountInString(d.Password) < 8 {
		errorsMap["password"] = append(errorsMap["password"], "too short, must be at least 8 characters")
	}

	if !emailRegex.MatchString(d.Email) {
		errorsMap["email"] = append(errorsMap["email"], "invalid email")
	}

	errors := make([]fieldError, 0, 4)
	fields := make([]string, 0, 4)
	for field, msgs := range errorsMap {
		if len(msgs) > 0 {
			errors = append(errors, fieldError{
				Field:    field,
				Messages: msgs,
			})
			fields = append(fields, field)
		}
	}

	if len(errors) > 0 {
		return ero.NewValidation(erolog.NewContextBuilder().With("fields", fields).Build(), errors)
	}

	return nil
}
