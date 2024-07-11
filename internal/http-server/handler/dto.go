package handler

import (
	"time"

	"github.com/Onnywrite/tinkoff-prod/internal/models"
)

type profile struct {
	Id       uint64         `json:"id"`
	Name     string         `json:"name"`
	Lastname string         `json:"surname"`
	Email    string         `json:"email"`
	Country  models.Country `json:"country"`
	IsPublic bool           `json:"is_public"`
	Image    string         `json:"image"`
	Birthday string         `json:"birthday"`
}

func getProfile(user *models.User) profile {
	return profile{
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
