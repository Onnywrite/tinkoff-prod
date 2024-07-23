package users

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Onnywrite/tinkoff-prod/internal/lib/tokens"
	"github.com/Onnywrite/tinkoff-prod/internal/models"
	"github.com/Onnywrite/tinkoff-prod/pkg/ero"
	"github.com/Onnywrite/tinkoff-prod/pkg/erolog"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthorizedUser struct {
	Profile Profile `json:"profile"`
	tokens.Pair
}

type PrivateProfile struct {
	Id       uint64 `json:"id"`
	Name     string `json:"name"`
	Lastname string `json:"surname"`
	IsPublic bool   `json:"is_public"`
}

type Profile struct {
	Id       uint64         `json:"id"`
	Name     string         `json:"name"`
	Lastname string         `json:"surname"`
	Email    string         `json:"email"`
	Country  models.Country `json:"country"`
	IsPublic bool           `json:"is_public"`
	Image    string         `json:"image"`
	Birthday string         `json:"birthday"`
}

func GetPrivateProfile(user *models.User) PrivateProfile {
	return PrivateProfile{
		Id:       user.Id,
		Name:     user.Name,
		Lastname: user.Lastname,
		IsPublic: user.IsPublic,
	}
}

func GetProfile(user *models.User) Profile {
	return Profile{
		Id:       user.Id,
		Name:     user.Name,
		Lastname: user.Lastname,
		Email:    user.Email,
		Country:  user.Country,
		IsPublic: user.IsPublic,
		Image:    user.Image,
		Birthday: user.Birthday.Format(time.DateOnly),
	}
}

type RegisterData struct {
	Name      string   `json:"name"`
	Lastname  string   `json:"surname"`
	Email     string   `json:"email"`
	Image     string   `json:"image"`
	CountryId uint64   `json:"country_id"`
	IsPublic  *bool    `json:"is_public,omitempty"`
	Birthday  dateOnly `json:"birthday"`
	Password  string   `json:"password"`
}

func (d *RegisterData) Validate() ero.Error {
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

	if !emailRegex.MatchString(d.Email) {
		errorsMap["email"] = append(errorsMap["email"], "invalid email")
	}

	if utf8.RuneCountInString(d.Password) < 8 {
		errorsMap["password"] = append(errorsMap["password"], "too short, must be at least 8 characters")
	}

	if len(d.Password) > 72 {
		errorsMap["password"] = append(errorsMap["password"], "too long, must be less than or equals 72 characters")
	}

	if d.Image == "" {
		d.Image = "default.png"
	}
	if d.CountryId == 0 {
		d.CountryId = 70
	}
	if d.IsPublic == nil {
		d.IsPublic = new(bool)
		*d.IsPublic = true
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
